package observability

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// HTTPMiddleware creates middleware for HTTP request tracing and metrics.
func (p *Provider) HTTPMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		// Wrap with OpenTelemetry HTTP instrumentation
		handler := otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Start timing the request
			start := time.Now()

			// Increment active requests
			p.MetricsProvider.IncrementActiveRequests(r.Context())

			// Create a wrapped response writer to capture status code
			ww := &wrappedResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Get the route pattern for better cardinality
			routePattern := getRoutePattern(r)

			// Add span attributes
			span := trace.SpanFromContext(r.Context())
			if span.IsRecording() {
				span.SetAttributes(
					attribute.String("http.method", r.Method),
					attribute.String("http.url", r.URL.String()),
					attribute.String("http.route", routePattern),
					attribute.String("http.user_agent", r.UserAgent()),
					attribute.String("http.remote_addr", r.RemoteAddr),
				)
			}

			// Call the next handler
			next.ServeHTTP(ww, r)

			// Record metrics after request completion
			duration := time.Since(start)
			statusCode := strconv.Itoa(ww.statusCode)

			p.MetricsProvider.RecordHTTPRequest(
				r.Context(),
				r.Method,
				routePattern,
				statusCode,
				duration,
			)

			// Decrement active requests
			p.MetricsProvider.DecrementActiveRequests(r.Context())

			// Update span with response information
			if span.IsRecording() {
				span.SetAttributes(
					attribute.Int("http.status_code", ww.statusCode),
					attribute.Int64("http.response_size", ww.bytesWritten),
				)

				if ww.statusCode >= 400 {
					span.SetStatus(codes.Error, http.StatusText(ww.statusCode))
				}
			}
		}), p.TracingProvider.config.ServiceName)

		return handler
	}
}

// wrappedResponseWriter wraps http.ResponseWriter to capture metrics.
type wrappedResponseWriter struct {
	http.ResponseWriter

	statusCode   int
	bytesWritten int64
}

func (w *wrappedResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *wrappedResponseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.bytesWritten += int64(n)

	return n, err
}

// getRoutePattern extracts the route pattern from the request context.
func getRoutePattern(r *http.Request) string {
	// Try to get the route pattern from chi context
	if rctx := chi.RouteContext(r.Context()); rctx != nil {
		return rctx.RoutePattern()
	}

	// Fallback to the URL path
	return r.URL.Path
}

// HealthCheckMiddleware provides observability for health check endpoints.
func (p *Provider) HealthCheckMiddleware() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := p.TracingProvider.StartSpan(r.Context(), "health_check")
		defer span.End()

		start := time.Now()

		// Perform health checks here
		// For now, just return 200 OK
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))

		// Record metrics
		duration := time.Since(start)
		p.MetricsProvider.RecordHTTPRequest(
			ctx,
			r.Method,
			"/health",
			"200",
			duration,
		)
	}
}

// DatabaseMiddleware wraps database operations with observability.
func (p *Provider) DatabaseMiddleware() func(operation string) func() {
	return func(operation string) func() {
		start := time.Now()
		ctx := context.Background()

		// Start a span for the database operation
		_, span := p.TracingProvider.StartSpan(ctx, "db."+operation)
		span.SetAttributes(
			attribute.String("db.operation", operation),
			attribute.String("db.system", "postgres"), // or sqlite
		)

		return func() {
			defer span.End()

			duration := time.Since(start)
			p.MetricsProvider.RecordDBQuery(ctx, operation, duration)
		}
	}
}
