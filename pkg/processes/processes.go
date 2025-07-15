package processes

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

// Manager takes care of starting processes and stopping them gracefully.
// Processes should be grouped in a ProcessGroup, which can be added to the Manager
// Processes within a group should be independent, meaning they can be started and stopped asynchronously.
// Groups will be started in the same order they were added, and will be stopped in the reverse order.
//
// How to use:
//  1. Create new Manager via NewProcessManager()
//  2. Create new ProcessGroups
//     Add processes to the ProcessGroup, using the correct Add function.
//     NOTE: Processes should NOT already have started operating, the start of the process should be defined in the start function.
//     Function-calls to other processes should NOT be used before start, since other components have also not started yet.
//  3. Add the ProcessGroup to the Manager
//  4. Once all processes are added to ProcessGroups, which are added to the Manager, call Run().
//     Run() will start all ProcessGroups in order and wil wait until a shutdownSignal is received to stop the groups.
type Manager struct {
	logger        *slog.Logger
	shutdownCtx   context.Context
	contextGroups []*ProcessGroup
}

func NewManager(logger *slog.Logger) *Manager {
	shutdownCtx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	cm := &Manager{
		logger:        logger.With(slog.String("component", "process_manager")),
		shutdownCtx:   shutdownCtx,
		contextGroups: make([]*ProcessGroup, 0),
	}

	logger.Info("process manager initialized", slog.Int("max_groups", cap(cm.contextGroups)))
	return cm
}

func (c *Manager) AddProcessGroup(group *ProcessGroup) {
	c.logger.Debug("adding process group to manager",
		slog.String("group_name", group.name),
		slog.Int("current_groups", len(c.contextGroups)))
	c.contextGroups = append(c.contextGroups, group)
	c.logger.Info("process group added successfully",
		slog.String("group_name", group.name),
		slog.Int("total_groups", len(c.contextGroups)))
}

// Run starts all groups in order in which they were added.
// If a start fails, the already started groups (including the failing one) are stopped and the error is returned.
// If all starts succeed, Run waits until:
//   - a shutDown signal is received
//   - an error is read from the errChan (error occurred in a process)
//   - the errChan is closed (all processes finished)
//
// and calls the stop functions of the groups, in reverse order in which they were added.
// If an error occurred during startup or shutdown, the error is returned.
func (c *Manager) Run() error {
	c.logger.Info("starting process manager",
		slog.Int("total_groups", len(c.contextGroups)))

	errChan := make(chan error, 20)
	var globalErr error

	// Start all process groups in order
	for i, group := range c.contextGroups {
		c.logger.Debug("starting process group",
			slog.String("group_name", group.name),
			slog.Int("group_index", i),
			slog.Int("remaining", len(c.contextGroups)-i-1))

		err := group.Start(c.logger, errChan)
		if err != nil {
			c.logger.Error("failed to start process group",
				slog.String("group_name", group.name),
				slog.Int("group_index", i),
				slog.String("error", err.Error()))
			globalErr = err
			break
		}

		c.logger.Info("process group started successfully",
			slog.String("group_name", group.name),
			slog.Int("group_index", i))
	}

	if globalErr == nil {
		c.logger.Info("all process groups started successfully",
			slog.Int("total_groups", len(c.contextGroups)))
	} else {
		c.logger.Warn("startup failed, will proceed to shutdown",
			slog.String("error", globalErr.Error()))
	}

	// Wait for shutdown signal or error
	select {
	case <-c.shutdownCtx.Done():
		c.logger.Info("shutdown signal received, stopping all process groups")
	case err, ok := <-errChan:
		if ok {
			c.logger.Error("error received from process group, initiating shutdown",
				slog.String("error", err.Error()))
			globalErr = err
		} else {
			c.logger.Info("all processes finished, shutting down")
		}
	}

	// Stop all process groups in reverse order
	c.logger.Info("stopping process groups",
		slog.Int("total_groups", len(c.contextGroups)))

	var i int
	var stopErr error
	for i = len(c.contextGroups) - 1; i >= 0; i-- {
		group := c.contextGroups[i]
		c.logger.Debug("stopping process group",
			slog.String("group_name", group.name),
			slog.Int("group_index", i))

		err := group.Stop(c.logger)
		if err != nil {
			c.logger.Error("failed to stop process group",
				slog.String("group_name", group.name),
				slog.Int("group_index", i),
				slog.String("error", err.Error()))
			stopErr = err
		} else {
			c.logger.Info("process group stopped successfully",
				slog.String("group_name", group.name),
				slog.Int("group_index", i))
		}
	}

	if stopErr != nil {
		c.logger.Error("process manager shutdown completed with errors",
			slog.String("stop_error", stopErr.Error()))
		if globalErr != nil {
			c.logger.Error("startup error was also present",
				slog.String("startup_error", globalErr.Error()))
		}
		return stopErr
	}

	if globalErr != nil {
		c.logger.Warn("process manager shutdown completed, but startup had failed",
			slog.String("startup_error", globalErr.Error()))
	} else {
		c.logger.Info("process manager shutdown completed successfully")
	}

	return globalErr // always return error to still signal k8s the stop was unplanned.
}
