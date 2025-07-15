package disruptor

import (
	"github.com/rs/zerolog"
)

type Option func(*Options)

type Options struct {
	logger   zerolog.Logger
	commands []Command
	handlers []interface{}
}

func DefaultOptions() *Options {
	return &Options{
		logger:   zerolog.Nop(),
		commands: []Command{},
	}
}

func WithCommands(cmds ...Command) Option {
	return func(session *Options) {
		for _, cmd := range cmds {
			if cmd == nil {
				continue // Skip nil commands
			}
			session.commands = append(session.commands, cmd)
		}
	}
}

func WithHandlers(handlers ...interface{}) Option {
	return func(session *Options) {
		for _, handler := range handlers {
			if handler == nil {
				continue // Skip nil handlers
			}
			session.handlers = append(session.handlers, handler)
		}
	}
}
