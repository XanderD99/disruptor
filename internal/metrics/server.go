package metrics

import (
	"context"
	"fmt"
	nethttp "net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/XanderD99/discord-disruptor/internal/http"
)

const (
	_pathMetrics = "/metrics"
)

type Config struct {
	// ⏳ How long to wait before shutting down the metrics server
	ShutdownDuration time.Duration `env:"SHUTDOWN_DURATION" default:"15s"`
	// 📊 Port where the metrics server will be available
	Port int `env:"PORT" default:"9090"`
}

type Server struct {
	cfg    Config
	server *http.Server
}

func NewServer(cfg Config) (*Server, error) {
	routes := nethttp.NewServeMux()
	routes.Handle(_pathMetrics, promhttp.Handler())
	s, err := http.NewServer(cfg.Port, routes)
	if err != nil {
		return nil, fmt.Errorf("creating new metrics server: %w", err)
	}

	return &Server{
		server: s,
		cfg:    cfg,
	}, nil
}

func (s *Server) Run(ctx context.Context) error {
	if err := s.server.Run(ctx, s.cfg.ShutdownDuration); err != nil {
		return fmt.Errorf("starting server: %w", err)
	}

	return nil
}
