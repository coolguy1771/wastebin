package observability

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// Config holds all observability configuration.
type Config struct {
	Tracing TracingConfig
	Metrics MetricsConfig
}

// Provider manages all observability components.
type Provider struct {
	TracingProvider *TracingProvider
	MetricsProvider *MetricsProvider
	logger          *zap.Logger
	startTime       time.Time
}

// New creates a new observability provider with the given configuration.
func New(config Config, logger *zap.Logger) (*Provider, error) {
	// Initialize tracing
	tracingProvider, err := NewTracingProvider(config.Tracing)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tracing: %w", err)
	}

	// Initialize metrics
	metricsProvider, err := NewMetricsProvider(
		config.Metrics,
		config.Tracing.ServiceName,
		config.Tracing.Version,
		config.Tracing.Environment,
	)
	if err != nil {
		// Cleanup tracing if metrics fail
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		tracingProvider.Shutdown(ctx)

		return nil, fmt.Errorf("failed to initialize metrics: %w", err)
	}

	provider := &Provider{
		TracingProvider: tracingProvider,
		MetricsProvider: metricsProvider,
		logger:          logger,
		startTime:       time.Now(),
	}

	// Start background goroutine to update system metrics
	go provider.updateSystemMetrics()

	logger.Info("Observability initialized",
		zap.Bool("tracing_enabled", config.Tracing.Enabled),
		zap.Bool("metrics_enabled", config.Metrics.Enabled),
		zap.String("service_name", config.Tracing.ServiceName),
		zap.String("service_version", config.Tracing.Version),
		zap.String("environment", config.Tracing.Environment),
	)

	return provider, nil
}

// Shutdown gracefully shuts down all observability components.
func (p *Provider) Shutdown(ctx context.Context) error {
	p.logger.Info("Shutting down observability")

	// Create a context with timeout for shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Shutdown metrics first
	err := p.MetricsProvider.Shutdown(shutdownCtx)
	if err != nil {
		p.logger.Error("Failed to shutdown metrics provider", zap.Error(err))
	}

	// Shutdown tracing
	err = p.TracingProvider.Shutdown(shutdownCtx)
	if err != nil {
		p.logger.Error("Failed to shutdown tracing provider", zap.Error(err))

		return err
	}

	p.logger.Info("Observability shutdown completed")

	return nil
}

// updateSystemMetrics runs in a background goroutine to update system-level metrics.
func (p *Provider) updateSystemMetrics() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ctx := context.Background()
		uptime := time.Since(p.startTime)
		p.MetricsProvider.UpdateSystemUptime(ctx, uptime)
	}
}

// DefaultConfig returns a default observability configuration.
func DefaultConfig() Config {
	return Config{
		Tracing: TracingConfig{
			Enabled:     true,
			ServiceName: "wastebin",
			Version:     "1.0.0",
			Environment: "development",
			Endpoint:    "http://localhost:4318/v1/traces",
			Headers:     make(map[string]string),
		},
		Metrics: MetricsConfig{
			Enabled:  true,
			Endpoint: "http://localhost:4318/v1/metrics",
			Headers:  make(map[string]string),
			Interval: 15 * time.Second,
		},
	}
}
