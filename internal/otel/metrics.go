package otel

import (
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
)

// InitMetrics initializes OpenTelemetry metrics with Prometheus exporter
func InitMetrics() (*prometheus.Exporter, error) {
	// Create Prometheus exporter
	promExporter, err := prometheus.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create Prometheus exporter: %w", err)
	}

	// Create metric provider with Prometheus exporter
	meterProvider := metric.NewMeterProvider(
		metric.WithReader(promExporter),
	)

	// Set global meter provider
	otel.SetMeterProvider(meterProvider)

	return promExporter, nil
}
