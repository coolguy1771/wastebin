package observability

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/resource"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
)

// MetricsConfig holds configuration for metrics collection.
type MetricsConfig struct {
	Enabled  bool   `koanf:"METRICS_ENABLED"`
	Endpoint string `koanf:"OTLP_METRICS_ENDPOINT"`
	Headers  map[string]string
	Interval time.Duration `koanf:"METRICS_INTERVAL"`
}

// MetricsProvider manages the OpenTelemetry metrics setup.
type MetricsProvider struct {
	meter    metric.Meter
	provider *sdkmetric.MeterProvider
	config   MetricsConfig

	// Application-specific metrics
	HTTPRequestDuration metric.Float64Histogram
	HTTPRequestsTotal   metric.Int64Counter
	HTTPActiveRequests  metric.Int64UpDownCounter
	DBConnectionsActive metric.Int64UpDownCounter
	DBQueryDuration     metric.Float64Histogram
	PasteCreatedTotal   metric.Int64Counter
	PasteViewedTotal    metric.Int64Counter
	PasteDeletedTotal   metric.Int64Counter
	SystemUptime        metric.Float64Gauge
}

// NewMetricsProvider creates a new metrics provider with the given configuration.
func NewMetricsProvider(config MetricsConfig, serviceName, version, environment string) (*MetricsProvider, error) {
	if !config.Enabled {
		// Return a no-op provider if metrics are disabled
		return &MetricsProvider{
			meter:    otel.Meter("noop"),
			provider: nil,
			config:   config,
		}, nil
	}

	// Create resource with service information
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(version),
			semconv.DeploymentEnvironmentName(environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create OTLP metrics exporter
	// Note: WithEndpoint expects host:port format, not a full URL
	exporter, err := otlpmetrichttp.New(
		context.Background(),
		otlpmetrichttp.WithEndpoint(config.Endpoint),
		otlpmetrichttp.WithHeaders(config.Headers),
		otlpmetrichttp.WithInsecure(), // Use TLS in production
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics exporter: %w", err)
	}

	// Create metrics provider
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(config.Interval))),
		sdkmetric.WithResource(res),
	)

	// Set global metrics provider
	otel.SetMeterProvider(mp)

	// Create meter
	meter := otel.Meter(serviceName)

	// Initialize application-specific metrics
	provider := &MetricsProvider{
		meter:               meter,
		provider:            mp,
		config:              config,
		HTTPRequestDuration: nil,
		HTTPRequestsTotal:   nil,
		HTTPActiveRequests:  nil,
		DBConnectionsActive: nil,
		DBQueryDuration:     nil,
		PasteCreatedTotal:   nil,
		PasteViewedTotal:    nil,
		PasteDeletedTotal:   nil,
		SystemUptime:        nil,
	}

	initErr := provider.initializeMetrics()
	if initErr != nil {
		return nil, fmt.Errorf("failed to initialize metrics: %w", initErr)
	}

	return provider, nil
}

// initializeMetrics creates all the application-specific metrics.
func (mp *MetricsProvider) initializeMetrics() error {
	var err error

	// HTTP metrics
	mp.HTTPRequestDuration, err = mp.meter.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("Duration of HTTP requests in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return fmt.Errorf("failed to create http_request_duration_seconds histogram: %w", err)
	}

	mp.HTTPRequestsTotal, err = mp.meter.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
	)
	if err != nil {
		return fmt.Errorf("failed to create http_requests_total counter: %w", err)
	}

	mp.HTTPActiveRequests, err = mp.meter.Int64UpDownCounter(
		"http_active_requests",
		metric.WithDescription("Number of active HTTP requests"),
	)
	if err != nil {
		return fmt.Errorf("failed to create http_active_requests gauge: %w", err)
	}

	// Database metrics
	mp.DBConnectionsActive, err = mp.meter.Int64UpDownCounter(
		"db_connections_active",
		metric.WithDescription("Number of active database connections"),
	)
	if err != nil {
		return fmt.Errorf("failed to create db_connections_active gauge: %w", err)
	}

	mp.DBQueryDuration, err = mp.meter.Float64Histogram(
		"db_query_duration_seconds",
		metric.WithDescription("Duration of database queries in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return fmt.Errorf("failed to create db_query_duration_seconds histogram: %w", err)
	}

	// Application-specific metrics
	mp.PasteCreatedTotal, err = mp.meter.Int64Counter(
		"paste_created_total",
		metric.WithDescription("Total number of pastes created"),
	)
	if err != nil {
		return fmt.Errorf("failed to create paste_created_total counter: %w", err)
	}

	mp.PasteViewedTotal, err = mp.meter.Int64Counter(
		"paste_viewed_total",
		metric.WithDescription("Total number of pastes viewed"),
	)
	if err != nil {
		return fmt.Errorf("failed to create paste_viewed_total counter: %w", err)
	}

	mp.PasteDeletedTotal, err = mp.meter.Int64Counter(
		"paste_deleted_total",
		metric.WithDescription("Total number of pastes deleted"),
	)
	if err != nil {
		return fmt.Errorf("failed to create paste_deleted_total counter: %w", err)
	}

	mp.SystemUptime, err = mp.meter.Float64Gauge(
		"system_uptime_seconds",
		metric.WithDescription("System uptime in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return fmt.Errorf("failed to create system_uptime_seconds gauge: %w", err)
	}

	return nil
}

// Meter returns the OpenTelemetry meter.
//
//nolint:ireturn // OpenTelemetry API requires returning an interface.
func (mp *MetricsProvider) Meter() metric.Meter {
	return mp.meter
}

// Shutdown gracefully shuts down the metrics provider.
func (mp *MetricsProvider) Shutdown(ctx context.Context) error {
	if mp.provider != nil {
		err := mp.provider.Shutdown(ctx)
		if err != nil {
			return fmt.Errorf("failed to shutdown metrics provider: %w", err)
		}
	}

	return nil
}

// Convenience methods for common metrics operations

// RecordHTTPRequest records metrics for an HTTP request.
func (mp *MetricsProvider) RecordHTTPRequest(ctx context.Context, method, path, status string, duration time.Duration) {
	if mp.config.Enabled {
		attributes := metric.WithAttributes(
			attribute.String("method", method),
			attribute.String("path", path),
			attribute.String("status", status),
		)

		mp.HTTPRequestDuration.Record(ctx, duration.Seconds(), attributes)
		mp.HTTPRequestsTotal.Add(ctx, 1, attributes)
	}
}

// IncrementActiveRequests increments the active HTTP requests counter.
func (mp *MetricsProvider) IncrementActiveRequests(ctx context.Context) {
	if mp.config.Enabled {
		mp.HTTPActiveRequests.Add(ctx, 1)
	}
}

// DecrementActiveRequests decrements the active HTTP requests counter.
func (mp *MetricsProvider) DecrementActiveRequests(ctx context.Context) {
	if mp.config.Enabled {
		mp.HTTPActiveRequests.Add(ctx, -1)
	}
}

// RecordDBQuery records metrics for a database query.
func (mp *MetricsProvider) RecordDBQuery(ctx context.Context, operation string, duration time.Duration) {
	if mp.config.Enabled {
		attributes := metric.WithAttributes(
			attribute.String("operation", operation),
		)
		mp.DBQueryDuration.Record(ctx, duration.Seconds(), attributes)
	}
}

// UpdateDBConnections updates the active database connections gauge.
func (mp *MetricsProvider) UpdateDBConnections(ctx context.Context, count int64) {
	if mp.config.Enabled {
		mp.DBConnectionsActive.Add(ctx, count)
	}
}

// RecordPasteCreated increments the paste created counter.
func (mp *MetricsProvider) RecordPasteCreated(ctx context.Context, language string) {
	if mp.config.Enabled {
		attributes := metric.WithAttributes(
			attribute.String("language", language),
		)
		mp.PasteCreatedTotal.Add(ctx, 1, attributes)
	}
}

// RecordPasteViewed increments the paste viewed counter.
func (mp *MetricsProvider) RecordPasteViewed(ctx context.Context, language string) {
	if mp.config.Enabled {
		attributes := metric.WithAttributes(
			attribute.String("language", language),
		)
		mp.PasteViewedTotal.Add(ctx, 1, attributes)
	}
}

// RecordPasteDeleted increments the paste deleted counter.
func (mp *MetricsProvider) RecordPasteDeleted(ctx context.Context, reason string) {
	if mp.config.Enabled {
		attributes := metric.WithAttributes(
			attribute.String("reason", reason),
		)
		mp.PasteDeletedTotal.Add(ctx, 1, attributes)
	}
}

// UpdateSystemUptime updates the system uptime gauge.
func (mp *MetricsProvider) UpdateSystemUptime(ctx context.Context, uptime time.Duration) {
	if mp.config.Enabled {
		mp.SystemUptime.Record(ctx, uptime.Seconds())
	}
}
