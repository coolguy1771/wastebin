# Wastebin Observability

This directory contains the observability configuration for the Wastebin application, including distributed tracing and metrics collection using OpenTelemetry.

## Components

### Distributed Tracing

- **OpenTelemetry SDK**: Integrated into the application for automatic instrumentation
- **Jaeger**: Distributed tracing backend for storing and visualizing traces
- **OTLP**: OpenTelemetry Protocol for sending telemetry data

### Metrics Collection

- **OpenTelemetry Metrics**: Custom metrics for business logic and system performance
- **Prometheus**: Time-series database for storing metrics
- **Grafana**: Visualization and alerting platform

### Data Processing

- **OpenTelemetry Collector**: Centralized telemetry data processing, routing, and transformation

## Quick Start

### 1. Start the Observability Stack

```bash
# Start all observability services
docker-compose -f docker-compose.observability.yml up -d

# Check service health
docker-compose -f docker-compose.observability.yml ps
```

### 2. Access the Dashboards

- **Grafana**: <http://localhost:3001> (admin/admin)
- **Jaeger UI**: <http://localhost:16686>
- **Prometheus**: <http://localhost:9090>

### 3. Configure Environment Variables

For local development, set these environment variables:

```bash
export WASTEBIN_TRACING_ENABLED=true
export WASTEBIN_METRICS_ENABLED=true
export WASTEBIN_SERVICE_NAME=wastebin
export WASTEBIN_SERVICE_VERSION=1.0.0
export WASTEBIN_ENVIRONMENT=development
export WASTEBIN_OTLP_TRACE_ENDPOINT=localhost:4318
export WASTEBIN_OTLP_METRICS_ENDPOINT=localhost:4318
export WASTEBIN_METRICS_INTERVAL=30
```

## Metrics

The application collects the following custom metrics:

### HTTP Metrics

- `http_request_duration_seconds`: Histogram of HTTP request durations
- `http_requests_total`: Counter of total HTTP requests by method, path, and status
- `http_active_requests`: Gauge of currently active HTTP requests

### Database Metrics

- `db_query_duration_seconds`: Histogram of database query durations by operation
- `db_connections_active`: Gauge of active database connections

### Application Metrics

- `paste_created_total`: Counter of pastes created by language
- `paste_viewed_total`: Counter of pastes viewed by language
- `paste_deleted_total`: Counter of pastes deleted by reason
- `system_uptime_seconds`: Gauge of system uptime

## Traces

The application automatically creates traces for:

- HTTP requests (with route patterns, status codes, response times)
- Database operations (with query types and durations)
- Business logic operations (paste creation, viewing, deletion)

### Trace Attributes

Common attributes added to spans:

- `service.name`: wastebin
- `service.version`: Application version
- `deployment.environment`: development/staging/production
- `http.method`: HTTP request method
- `http.route`: Route pattern (e.g., `/api/v1/paste/{uuid}`)
- `http.status_code`: HTTP response status code
- `db.operation`: Database operation type

## Configuration

### Application Configuration

The observability configuration is managed through environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `WASTEBIN_TRACING_ENABLED` | `true` | Enable distributed tracing |
| `WASTEBIN_METRICS_ENABLED` | `true` | Enable metrics collection |
| `WASTEBIN_SERVICE_NAME` | `wastebin` | Service name for telemetry |
| `WASTEBIN_SERVICE_VERSION` | `1.0.0` | Service version |
| `WASTEBIN_ENVIRONMENT` | `development` | Deployment environment |
| `WASTEBIN_OTLP_TRACE_ENDPOINT` | `localhost:4318` | OTLP traces endpoint (host:port) |
| `WASTEBIN_OTLP_METRICS_ENDPOINT` | `localhost:4318` | OTLP metrics endpoint (host:port) |
| `WASTEBIN_METRICS_INTERVAL` | `30` | Metrics export interval (seconds) |

### Production Deployment

For production deployments:

1. **Use TLS**: Configure TLS for all OTLP endpoints
2. **Authentication**: Set up proper authentication for exporters
3. **Sampling**: Configure trace sampling to reduce overhead
4. **Resource Limits**: Set appropriate resource limits for collector and storage
5. **Retention**: Configure appropriate data retention policies

Example production configuration:

```yaml
# Update docker-compose.observability.yml
services:
  wastebin:
    environment:
      - WASTEBIN_ENVIRONMENT=production
      - WASTEBIN_OTLP_TRACE_ENDPOINT=your-jaeger-instance:4318
      - WASTEBIN_OTLP_METRICS_ENDPOINT=your-prometheus-gateway:4318
```

## Monitoring and Alerting

### Grafana Dashboards

Pre-configured dashboards are available:

1. **Wastebin Overview**: Application performance metrics
2. **HTTP Performance**: Request rates, response times, error rates
3. **Database Performance**: Query performance and connection metrics
4. **Business Metrics**: Paste operations and usage patterns

### Recommended Alerts

Set up alerts for:

- High HTTP error rates (>1%)
- High response times (>1s 95th percentile)
- Database connection exhaustion
- Application downtime
- High memory/CPU usage

## Troubleshooting

### No Traces Appearing

1. Check OTLP endpoint configuration
2. Verify network connectivity to collector
3. Check collector logs: `docker-compose logs otel-collector`
4. Verify trace sampling configuration

### No Metrics in Prometheus

1. Check metrics endpoint configuration
2. Verify Prometheus scrape configuration
3. Check collector metrics pipeline
4. Verify metric export interval

### High Overhead

1. Reduce trace sampling rate
2. Increase metrics export interval
3. Configure batch processors in collector
4. Review resource allocations

## Development

### Adding New Metrics

```go
// In handlers or business logic
func (h *Handler) CreatePaste(w http.ResponseWriter, r *http.Request) {
    // Record custom metrics
    h.observability.MetricsProvider.RecordPasteCreated(r.Context(), language)
    
    // Add trace information
    span := trace.SpanFromContext(r.Context())
    span.SetAttributes(attribute.String("paste.language", language))
}
```

### Adding New Traces

```go
// Create custom spans
ctx, span := h.observability.TracingProvider.StartSpan(ctx, "custom_operation")
defer span.End()

// Add custom attributes
span.SetAttributes(
    attribute.String("operation.type", "validation"),
    attribute.Int("paste.size", len(content)),
)
```

## Architecture

```
┌─────────────┐    ┌─────────────────┐    ┌──────────────┐
│  Wastebin   │───▶│ OTLP Collector  │───▶│   Jaeger     │
│ Application │    │                 │    │ (Traces)     │
└─────────────┘    └─────────────────┘    └──────────────┘
                            │
                            ▼
                   ┌──────────────┐    ┌──────────────┐
                   │ Prometheus   │───▶│   Grafana    │
                   │ (Metrics)    │    │(Visualization)│
                   └──────────────┘    └──────────────┘
```

The OpenTelemetry Collector acts as a central hub for all telemetry data, providing:

- Protocol translation
- Data transformation and enrichment
- Routing to multiple backends
- Buffering and retry logic
- Performance optimization
