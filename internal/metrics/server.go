package metrics

import (
	"context"
	"fmt"
	nethttp "net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/exporters/prometheus"

	"github.com/XanderD99/disruptor/internal/http"
)

const (
	_pathMetrics = "/metrics"
)

const (
	StatusSuccess = "success"
	StatusError   = "error"
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

// NewServerWithExporter creates a new metrics server using OpenTelemetry Prometheus exporter
func NewServerWithExporter(cfg Config, promExporter *prometheus.Exporter) (*Server, error) {
	routes := nethttp.NewServeMux()
	// OpenTelemetry Prometheus exporter should implement http.Handler
	// If not, we'll need to access its internal gatherer
	if handler, ok := any(promExporter).(nethttp.Handler); ok {
		routes.Handle(_pathMetrics, handler)
	} else {
		// Fallback to default prometheus handler
		routes.Handle(_pathMetrics, promhttp.Handler())
	}
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
