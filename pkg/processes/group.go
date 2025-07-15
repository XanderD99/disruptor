package processes

import (
	"context"
	"errors"
	"log/slog"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

type Process struct {
	name             string
	start            func() error
	startShouldBlock bool
	stop             func() error
}

// ProcessGroup is a group of processes which are independent of each other.
// They are started and stopped asynchronously.
// Processes should be added via the correct Add function.
// When the group stops, the stop function of all processes are called and the context is cancelled.
type ProcessGroup struct {
	name            string
	ctx             context.Context
	shutdown        context.CancelFunc
	shutdownTimeout time.Duration
	errGroup        *errgroup.Group
	started         bool
	processes       []Process
}

// NewGroup creates a new group.
// shutdownTimeout is the maximum a stop function can take, before the context is cancelled.
func NewGroup(name string, shutdownTimeout time.Duration) *ProcessGroup {
	ctx := context.Background()
	shutdownCtx, shutdown := signal.NotifyContext(ctx, syscall.SIGKILL) // must listen to at least one signal, otherwise all signals are listened to
	return &ProcessGroup{
		name:            name,
		ctx:             shutdownCtx,
		shutdown:        shutdown,
		shutdownTimeout: shutdownTimeout,
		errGroup:        &errgroup.Group{},
		processes:       make([]Process, 0),
	}
}

// AddProcess adds a process to the group.
// start will be called when the group is started, and optionally stop when the group is stopped.
// startShouldBlock indicates whether the group should wait for the start function to finish during startup,
// this should be false if the start function has an infinite loop.
func (p *ProcessGroup) AddProcess(name string, start func() error, startShouldBlock bool, stop func() error) {
	p.checkStart(name)
	p.processes = append(p.processes, Process{
		name:             name,
		start:            start,
		startShouldBlock: startShouldBlock,
		stop:             stop,
	})
}

// AddProcessWithCtx works similar to AddProcess, but also provides a context to the start function.
// This context will be canceled when the group is stopped, after first calling the stop function.
func (p *ProcessGroup) AddProcessWithCtx(name string, start func(ctx context.Context) error, startShouldBlock bool, stop func() error) {
	p.checkStart(name)
	p.processes = append(p.processes, Process{
		name:             name,
		start:            func() error { return start(p.ctx) },
		startShouldBlock: startShouldBlock,
		stop:             stop,
	})
}

// AddProcessWithoutStart works similar to AddProcess, but only requires the stop function.
func (p *ProcessGroup) AddProcessWithoutStart(name string, stop func() error) {
	p.checkStart(name)
	p.processes = append(p.processes, Process{
		name:  name,
		start: func() error { return nil },
		stop:  stop,
	})
}

// start calls the start function of every process.
// If the start of a process with startShouldBlock fails, an error is returned and added to the errChan.
// If a non-blocking start fails (could be after seconds/days/weeks), an error is added to the errChan.
func (p *ProcessGroup) Start(logger *slog.Logger, errChan chan error) error {
	groupLogger := logger.With(slog.String("group", p.name))
	groupLogger.Info("starting process group",
		slog.Int("total_processes", len(p.processes)),
		slog.String("shutdown_timeout", p.shutdownTimeout.String()))

	p.started = true

	for i, process := range p.processes {
		processLogger := groupLogger.With(
			slog.String("process", process.name),
			slog.Int("process_index", i),
			slog.Bool("should_block", process.startShouldBlock))

		if process.startShouldBlock {
			processLogger.Debug("starting blocking process")
			err := process.start()
			if err != nil {
				processLogger.Error("blocking process failed to start",
					slog.String("error", err.Error()))
				errChan <- err
				return err
			}
			processLogger.Info("blocking process started successfully")
		} else {
			processLogger.Debug("starting non-blocking process")
			p.errGroup.Go(func() error {
				if err := process.start(); err != nil {
					processLogger.Error("non-blocking process failed",
						slog.String("error", err.Error()))
					errChan <- err
				} else {
					processLogger.Info("non-blocking process started successfully")
				}
				return nil
			})
		}
	}

	groupLogger.Info("process group startup completed",
		slog.Int("started_processes", len(p.processes)))
	return nil
}

// stop calls the stop function of every process (if provided), after which the context of the group is cancelled.
// This function will first block until all stop functions are finished (with timeout),
// after which all processes which did not block during startup are finished.
func (p *ProcessGroup) Stop(logger *slog.Logger) error {
	groupLogger := logger.With(slog.String("group", p.name))

	if !p.started {
		groupLogger.Debug("process group was not started, skipping stop")
		return nil
	}

	groupLogger.Info("stopping process group",
		slog.Int("total_processes", len(p.processes)),
		slog.String("shutdown_timeout", p.shutdownTimeout.String()))

	var wg sync.WaitGroup
	var globalErr error
	stoppedProcesses := 0

	for i, process := range p.processes {
		if process.stop != nil {
			wg.Add(1)
			go func(proc Process, index int) {
				defer wg.Done()
				processLogger := groupLogger.With(
					slog.String("process", proc.name),
					slog.Int("process_index", index))

				processLogger.Debug("stopping process")
				err := proc.stop()
				if err != nil {
					processLogger.Error("failed to stop process",
						slog.String("error", err.Error()))
					globalErr = err
				} else {
					processLogger.Info("process stopped successfully")
				}
			}(process, i)
			stoppedProcesses++
		}
	}

	groupLogger.Debug("waiting for process stop functions to complete",
		slog.Int("processes_with_stop", stoppedProcesses))

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		groupLogger.Info("all process stop functions completed")
	case <-time.After(p.shutdownTimeout):
		groupLogger.Error("stopping of process group timed out",
			slog.String("timeout", p.shutdownTimeout.String()))
		globalErr = errors.New("stopping of process group timed out")
	}

	groupLogger.Debug("cancelling group context to stop remaining processes")
	p.shutdown() // Cancel context -> stop all processes which rely on the context for stopping.

	groupLogger.Debug("waiting for non-blocking processes to finish")
	err := p.errGroup.Wait() // Wait for processes which are stopped via context cancelled
	if err != nil {
		groupLogger.Error("error while waiting for non-blocking processes",
			slog.String("error", err.Error()))
		return err
	}

	if globalErr != nil {
		groupLogger.Error("process group stopped with errors",
			slog.String("error", globalErr.Error()))
	} else {
		groupLogger.Info("process group stopped successfully")
	}

	return globalErr
}

func (p *ProcessGroup) checkStart(name string) {
	if p.started {
		// This is a programming error and should panic in development
		// but we'll just log it as error in production to avoid crashing
		// TODO: Consider making this configurable based on environment
		panic("cannot add process '" + name + "' to group '" + p.name + "' after it has been started")
	}
}
