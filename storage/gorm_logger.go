package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	gormlogger "gorm.io/gorm/logger"

	"github.com/coolguy1771/wastebin/log"
)

const defaultSlowThreshold = 200 * time.Millisecond

// GormZapLogger implements the GORM logger interface using Zap.
type GormZapLogger struct {
	logger               *log.Logger
	SlowThreshold        time.Duration
	LogLevel             gormlogger.LogLevel
	IgnoreRecordNotFound bool
	ParameterizedQueries bool
	Colorful             bool
}

// NewGormZapLogger creates a new GORM logger that uses Zap.
func NewGormZapLogger(zapLogger *log.Logger) *GormZapLogger {
	return &GormZapLogger{
		logger:               zapLogger,
		SlowThreshold:        defaultSlowThreshold,
		LogLevel:             gormlogger.Warn,
		IgnoreRecordNotFound: true,
		ParameterizedQueries: false,
		Colorful:             false,
	}
}

// LogMode sets the log level.
//
//nolint:ireturn // Required to implement gorm's logger interface
func (l *GormZapLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

// Info logs info messages.
func (l *GormZapLogger) Info(_ context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Info {
		l.logger.Info(fmt.Sprintf(msg, data...))
	}
}

// Warn logs warning messages.
func (l *GormZapLogger) Warn(_ context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Warn {
		l.logger.Warn(fmt.Sprintf(msg, data...))
	}
}

// Error logs error messages.
func (l *GormZapLogger) Error(_ context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Error {
		l.logger.Error(fmt.Sprintf(msg, data...))
	}
}

// Trace logs SQL queries with execution time.
func (l *GormZapLogger) Trace(
	_ context.Context,
	begin time.Time,
	fc func() (sql string, rowsAffected int64),
	err error,
) {
	if l.LogLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	switch {
	case err != nil && l.LogLevel >= gormlogger.Error && (!errors.Is(err, gorm.ErrRecordNotFound) || !l.IgnoreRecordNotFound):
		l.logger.Error("GORM query error",
			zap.Error(err),
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= gormlogger.Warn:
		l.logger.Warn("GORM slow query",
			zap.Duration("elapsed", elapsed),
			zap.Duration("threshold", l.SlowThreshold),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)
	case l.LogLevel >= gormlogger.Info:
		l.logger.Debug("GORM query",
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)
	}
}
