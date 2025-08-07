package observability

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"

	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// TracingConfig holds configuration for distributed tracing.
type TracingConfig struct {
	Enabled     bool   `koanf:"TRACING_ENABLED"`
	ServiceName string `koanf:"SERVICE_NAME"`
	Version     string `koanf:"SERVICE_VERSION"`
	Environment string `koanf:"ENVIRONMENT"`
	Endpoint    string `koanf:"OTLP_TRACE_ENDPOINT"`
	Headers     map[string]string
}

// TracingProvider manages the OpenTelemetry tracing setup.
type TracingProvider struct {
	tracer   oteltrace.Tracer
	provider *trace.TracerProvider
	config   TracingConfig
}

// NewTracingProvider creates a new tracing provider with the given configuration.
func NewTracingProvider(config TracingConfig) (*TracingProvider, error) {
	if !config.Enabled {
		// Return a no-op provider if tracing is disabled
		return &TracingProvider{
			tracer:   otel.Tracer("noop"),
			provider: nil,
			config:   config,
		}, nil
	}

	// Create resource with service information
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.Version),
			semconv.DeploymentEnvironmentName(config.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create OTLP trace exporter
	// Note: WithEndpoint expects host:port format, not a full URL
	exporter, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpoint(config.Endpoint),
		otlptracehttp.WithHeaders(config.Headers),
		otlptracehttp.WithInsecure(), // Use TLS in production
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Create trace provider
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
		trace.WithSampler(trace.AlwaysSample()), // Configure sampling for production
	)

	// Set global trace provider
	otel.SetTracerProvider(tp)

	// Set global propagator for context propagation
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Create tracer
	tracer := otel.Tracer(config.ServiceName)

	return &TracingProvider{
		tracer:   tracer,
		provider: tp,
		config:   config,
	}, nil
}

// Tracer returns the OpenTelemetry tracer.
//
//nolint:ireturn // OpenTelemetry API requires returning an interface.
func (tp *TracingProvider) Tracer() oteltrace.Tracer {
	return tp.tracer
}

// Shutdown gracefully shuts down the tracing provider.
func (tp *TracingProvider) Shutdown(ctx context.Context) error {
	if tp.provider != nil {
		err := tp.provider.Shutdown(ctx)
		if err != nil {
			return fmt.Errorf("failed to shutdown tracing provider: %w", err)
		}
	}

	return nil
}

// StartSpan is a convenience method to start a new span.
//
//nolint:ireturn,spancheck // OpenTelemetry API requires returning an interface, and the span is ended by the caller.
func (tp *TracingProvider) StartSpan(
	ctx context.Context,
	name string,
	opts ...oteltrace.SpanStartOption,
) (context.Context, oteltrace.Span) {
	spanCtx, span := tp.tracer.Start(ctx, name, opts...)

	return spanCtx, span
}

// GetSpanFromContext retrieves the current span from context.
//
//nolint:ireturn // OpenTelemetry API requires returning an interface.
func GetSpanFromContext(ctx context.Context) oteltrace.Span {
	return oteltrace.SpanFromContext(ctx)
}

// AddSpanEvent adds an event to the current span.
func AddSpanEvent(ctx context.Context, name string, attributes ...oteltrace.EventOption) {
	span := oteltrace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.AddEvent(name, attributes...)
	}
}

// AddSpanError records an error in the current span.
func AddSpanError(ctx context.Context, err error) {
	span := oteltrace.SpanFromContext(ctx)
	if span.IsRecording() && err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

// SetSpanAttributes sets attributes on the current span.
func SetSpanAttributes(ctx context.Context, attributes ...attribute.KeyValue) {
	span := oteltrace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(attributes...)
	}
}
