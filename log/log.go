package log

import (
	"errors"
	"fmt"
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger to provide a simplified logging interface.
type Logger struct {
	l     *zap.Logger // zap ensure that zap.Logger is safe for concurrent use
	level zapcore.Level
}

// Debug logs a debug message with optional fields.
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.l.Debug(msg, fields...)
}

// Info logs an info message with optional fields.
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.l.Info(msg, fields...)
}

// Warn logs a warning message with optional fields.
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.l.Warn(msg, fields...)
}

// Error logs an error message with optional fields.
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.l.Error(msg, fields...)
}

// DPanic logs a panic message with optional fields in development mode.
func (l *Logger) DPanic(msg string, fields ...zap.Field) {
	l.l.DPanic(msg, fields...)
}

// Panic logs a panic message with optional fields.
func (l *Logger) Panic(msg string, fields ...zap.Field) {
	l.l.Panic(msg, fields...)
}

// Fatal logs a fatal message with optional fields.
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.l.Fatal(msg, fields...)
}

// Function variables for all field types in github.com/uber-go/zap/field.go.
//
//nolint:gochecknoglobals // These are function variables from the zap library, not global state.
var (
	// Skip creates a field that does nothing.
	Skip = zap.Skip
	// Binary creates a field that carries an opaque binary blob.
	Binary = zap.Binary
	// Bool creates a field that carries a bool.
	Bool = zap.Bool
	// Boolp creates a field that carries a *bool.
	Boolp = zap.Boolp
	// ByteString creates a field that carries UTF-8 encoded text as a []byte.
	ByteString = zap.ByteString
	// Float64 creates a field that carries a float64.
	Float64 = zap.Float64
	// Float64p creates a field that carries a *float64.
	Float64p = zap.Float64p
	// Float32 creates a field that carries a float32.
	Float32 = zap.Float32
	// Float32p creates a field that carries a *float32.
	Float32p = zap.Float32p
	// Durationp creates a field that carries a *time.Duration.
	Durationp = zap.Durationp
	// Any takes a key and an arbitrary value and chooses the best way to represent them as a field.
	Any = zap.Any
	// Object creates a field that carries an ObjectMarshaler.
	Object = zap.Object
)

// Default logger instance.
//
//nolint:gochecknoglobals // This is the default logger instance, which is a singleton.
var defaultLogger *Logger

// Setup initializes the default logger.
func Setup() {
	var err error

	defaultLogger, err = New(os.Stdout, "INFO")
	if err != nil {
		panic(fmt.Sprintf("failed to initialize default logger: %v", err))
	}
}

// Default returns the default logger instance.
func Default() *Logger {
	return defaultLogger
}

// New creates a new Logger with the specified writer and log level.
func New(writer io.Writer, level string) (*Logger, error) {
	parsedAtomicLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		return nil, fmt.Errorf("failed to parse log level %q: %w", level, err)
	}

	if writer == nil {
		return nil, errors.New("writer cannot be nil")
	}

	cfg := zap.NewProductionConfig()
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(cfg.EncoderConfig),
		zapcore.AddSync(writer),
		parsedAtomicLevel,
	)
	logger := &Logger{
		l:     zap.New(core),
		level: parsedAtomicLevel,
	}

	return logger, nil
}

// Sync flushes any buffered log entries.
func (l *Logger) Sync() error {
	err := l.l.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync logger: %w", err)
	}

	return nil
}

// ZapLogger returns the underlying zap.Logger.
func (l *Logger) ZapLogger() *zap.Logger {
	return l.l
}

// ResetDefault sets the default logger to the provided logger.
// Not safe for concurrent use.
func ResetDefault(l *Logger) {
	defaultLogger = l
}

// Sync flushes any buffered log entries from the default logger.
func Sync() error {
	if defaultLogger != nil {
		return defaultLogger.Sync()
	}

	return nil
}

// Debug logs a debug message using the default logger.
func Debug(msg string, fields ...zap.Field) {
	if defaultLogger == nil {
		return
	}
	defaultLogger.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	if defaultLogger == nil {
		return
	}
	defaultLogger.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	if defaultLogger == nil {
		return
	}
	defaultLogger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	if defaultLogger == nil {
		return
	}
	defaultLogger.Error(msg, fields...)
}

func DPanic(msg string, fields ...zap.Field) {
	if defaultLogger == nil {
		return
	}
	defaultLogger.DPanic(msg, fields...)
}

func Panic(msg string, fields ...zap.Field) {
	if defaultLogger == nil {
		return
	}
	defaultLogger.Panic(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	if defaultLogger == nil {
		return
	}
	defaultLogger.Fatal(msg, fields...)
}
