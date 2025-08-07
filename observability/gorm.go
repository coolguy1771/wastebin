package observability

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	// SQL operation detection constants.
	sqlPrefixLength = 10
	nanosecondsToMs = 1e6
)

// GormTracer implements gorm logger interface with OpenTelemetry tracing.
type GormTracer struct {
	logger.Interface

	provider *Provider
}

// NewGormTracer creates a new GORM tracer.
func NewGormTracer(provider *Provider, base logger.Interface) *GormTracer {
	return &GormTracer{
		Interface: base,
		provider:  provider,
	}
}

// Trace implements the gorm logger interface.
func (g *GormTracer) Trace(
	ctx context.Context,
	begin time.Time,
	functionCallback func() (sql string, rowsAffected int64),
	err error,
) {
	// Call the base logger first
	g.Interface.Trace(ctx, begin, functionCallback, err)

	// Only trace if provider is available and enabled
	if g.provider == nil || !g.provider.TracingProvider.config.Enabled {
		return
	}

	// Get SQL and rows affected
	sql, rowsAffected := functionCallback()

	// Determine operation type from SQL
	operation := extractOperationType(sql)

	// Create span
	ctx, span := g.provider.TracingProvider.StartSpan(ctx, "db."+operation)
	defer span.End()

	// Calculate duration
	duration := time.Since(begin)

	// Set span attributes
	span.SetAttributes(
		attribute.String("db.system", "postgres"), // or detect dynamically
		attribute.String("db.operation", operation),
		attribute.String("db.statement", sql),
		attribute.Int64("db.rows_affected", rowsAffected),
		attribute.Float64("db.duration_ms", float64(duration.Nanoseconds())/nanosecondsToMs),
	)

	// Record error if present
	if err != nil {
		AddSpanError(ctx, err)
	}

	// Record metrics
	g.provider.MetricsProvider.RecordDBQuery(ctx, operation, duration)
}

// extractOperationType extracts the operation type from SQL statement.
func extractOperationType(sql string) string {
	if sql == "" {
		return "unknown"
	}

	// Convert to lowercase for comparison
	sqlLower := sql
	if len(sql) > sqlPrefixLength {
		sqlLower = sql[:sqlPrefixLength]
	}

	// Simple operation detection
	switch {
	case contains(sqlLower, "SELECT"):
		return "select"
	case contains(sqlLower, "INSERT"):
		return "insert"
	case contains(sqlLower, "UPDATE"):
		return "update"
	case contains(sqlLower, "DELETE"):
		return "delete"
	case contains(sqlLower, "CREATE"):
		return "create"
	case contains(sqlLower, "DROP"):
		return "drop"
	case contains(sqlLower, "ALTER"):
		return "alter"
	default:
		return "query"
	}
}

// contains checks if a string contains a substring (case-insensitive).
func contains(str, substr string) bool {
	// Simple case-insensitive contains check
	if len(substr) > len(str) {
		return false
	}

	for i := 0; i <= len(str)-len(substr); i++ {
		match := true

		for j := range len(substr) {
			if toLower(str[i+j]) != toLower(substr[j]) {
				match = false

				break
			}
		}

		if match {
			return true
		}
	}

	return false
}

// toLower converts a byte to lowercase.
func toLower(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + ('a' - 'A')
	}

	return b
}

// InstrumentGorm adds observability to a GORM database instance.
func (p *Provider) InstrumentGorm(db *gorm.DB, baseLogger logger.Interface) *gorm.DB {
	if !p.TracingProvider.config.Enabled && !p.MetricsProvider.config.Enabled {
		return db
	}

	// Create new session with instrumented logger
	instrumentedDB := db.Session(&gorm.Session{
		Logger: NewGormTracer(p, baseLogger),
	})

	return instrumentedDB
}
