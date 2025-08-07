package config

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"

	"github.com/coolguy1771/wastebin/log"
)

// Configuration constants.
const (
	// Database configuration constants.
	DefaultDBMaxIdleConns = 10
	DefaultDBMaxOpenConns = 50
	DefaultDBPort         = 5432
	MaxPortNumber         = 65535
	MinPortNumber         = 1

	// Request size constants.
	DefaultMaxRequestSize = 15728640          // 15MB
	MinRequestSize        = 1024 * 1024       // 1MB
	MaxRequestSize        = 100 * 1024 * 1024 // 100MB

	// Metrics interval constants.
	DefaultMetricsInterval = 30 // seconds

	// Webapp port constants.
	DefaultWebappPort = "3000"
)

// Conf holds the global configuration instance.
//
//nolint:gochecknoglobals // Conf is a singleton that holds the application's configuration.
var Conf Config

// Config represents the overall configuration for the application.
type Config struct {
	// Database configuration
	DBUser         string `koanf:"DB_USER"`
	DBPassword     string `koanf:"DB_PASSWORD"`
	DBHost         string `koanf:"DB_HOST"`
	DBPort         int    `koanf:"DB_PORT"`
	DBName         string `koanf:"DB_NAME"`
	DBMaxIdleConns int    `koanf:"DB_MAX_IDLE_CONNS"`
	DBMaxOpenConns int    `koanf:"DB_MAX_OPEN_CONNS"`

	// Web application configuration
	WebappPort string `koanf:"WEBAPP_PORT"`
	Dev        bool   `koanf:"DEV"`
	LocalDB    bool   `koanf:"LOCAL_DB"`

	// TLS configuration
	TLSEnabled  bool   `koanf:"TLS_ENABLED"`
	TLSCertFile string `koanf:"TLS_CERT_FILE"`
	TLSKeyFile  string `koanf:"TLS_KEY_FILE"`

	// Security configuration
	AllowedOrigins string `koanf:"ALLOWED_ORIGINS"`
	RequireAuth    bool   `koanf:"REQUIRE_AUTH"`
	AuthUsername   string `koanf:"AUTH_USERNAME"`
	AuthPassword   string `koanf:"AUTH_PASSWORD"`
	CSRFKey        string `koanf:"CSRF_KEY"`
	MaxRequestSize int64  `koanf:"MAX_REQUEST_SIZE"`

	// Logger configuration
	LogLevel string `koanf:"LOG_LEVEL"`

	// Observability configuration
	TracingEnabled      bool   `koanf:"TRACING_ENABLED"`
	MetricsEnabled      bool   `koanf:"METRICS_ENABLED"`
	ServiceName         string `koanf:"SERVICE_NAME"`
	ServiceVersion      string `koanf:"SERVICE_VERSION"`
	Environment         string `koanf:"ENVIRONMENT"`
	OTLPTraceEndpoint   string `koanf:"OTLP_TRACE_ENDPOINT"`
	OTLPMetricsEndpoint string `koanf:"OTLP_METRICS_ENDPOINT"`
	MetricsInterval     int    `koanf:"METRICS_INTERVAL"` // in seconds
}

// Load initializes the application configuration by merging default values with
// environment variables, unmarshaling them into the global Config, validating
// the result, and returning a pointer to the loaded configuration.
// The application terminates if loading or validation fails.
func Load() *Config {
	koanfInstance := koanf.New(".")

	// Load default configuration settings
	err := koanfInstance.Load(confmap.Provider(map[string]interface{}{
		"WEBAPP_PORT":           DefaultWebappPort,
		"DB_MAX_IDLE_CONNS":     DefaultDBMaxIdleConns,
		"DB_MAX_OPEN_CONNS":     DefaultDBMaxOpenConns,
		"DB_PORT":               DefaultDBPort,
		"DB_HOST":               "localhost",
		"DB_USER":               "wastebin",
		"DB_NAME":               "wastebin",
		"LOG_LEVEL":             "INFO",
		"LOCAL_DB":              false,
		"TLS_ENABLED":           false,
		"TLS_CERT_FILE":         "",
		"TLS_KEY_FILE":          "",
		"ALLOWED_ORIGINS":       "",
		"REQUIRE_AUTH":          false,
		"AUTH_USERNAME":         "",
		"AUTH_PASSWORD":         "",
		"CSRF_KEY":              "",
		"MAX_REQUEST_SIZE":      DefaultMaxRequestSize,
		"TRACING_ENABLED":       true,
		"METRICS_ENABLED":       true,
		"SERVICE_NAME":          "wastebin",
		"SERVICE_VERSION":       "1.0.0",
		"ENVIRONMENT":           "development",
		"OTLP_TRACE_ENDPOINT":   "localhost:4318",
		"OTLP_METRICS_ENDPOINT": "localhost:4318",
		"METRICS_INTERVAL":      DefaultMetricsInterval,
	}, "."), nil)
	if err != nil {
		log.Error("Error loading default config", zap.Error(err))
	}

	// Load environment variables with prefix WASTEBIN_
	err = koanfInstance.Load(env.Provider("WASTEBIN_", ".", func(s string) string {
		return strings.TrimPrefix(s, "WASTEBIN_")
	}), nil)
	if err != nil {
		log.Error("Error loading environment config", zap.Error(err))
	}

	// Unmarshal the configuration into the Config struct
	err = koanfInstance.Unmarshal("", &Conf)
	if err != nil {
		log.Error("Error loading config", zap.Error(err))
	}

	// Validate the configuration
	err = Conf.Validate()
	if err != nil {
		log.Error("Configuration validation failed", zap.Error(err))
	}

	return &Conf
}

// Validate validates the configuration settings.
func (c *Config) Validate() error {
	var errs []string

	// Validate webapp port
	err := c.validateWebappPort()
	if err != nil {
		errs = append(errs, err.Error())
	}

	// Validate database configuration
	err = c.validateDatabaseConfig()
	if err != nil {
		errs = append(errs, err.Error())
	}

	// Validate connection pool settings
	err = c.validateConnectionPool()
	if err != nil {
		errs = append(errs, err.Error())
	}

	// Validate log level
	err = c.validateLogLevel()
	if err != nil {
		errs = append(errs, err.Error())
	}

	// Validate TLS configuration
	err = c.validateTLSConfig()
	if err != nil {
		errs = append(errs, err.Error())
	}

	// Validate authentication configuration
	err = c.validateAuthConfig()
	if err != nil {
		errs = append(errs, err.Error())
	}

	// Validate request size limits
	err = c.validateRequestSize()
	if err != nil {
		errs = append(errs, err.Error())
	}

	if len(errs) > 0 {
		return errors.New("configuration validation errors: " + strings.Join(errs, "; "))
	}

	return nil
}

// validateWebappPort validates the webapp port configuration.
func (c *Config) validateWebappPort() error {
	if c.WebappPort == "" {
		return errors.New("webapp port must be specified")
	}

	port, err := strconv.Atoi(c.WebappPort)
	if err != nil || port < MinPortNumber || port > MaxPortNumber {
		return errors.New("webapp port must be a valid port number (1-65535)")
	}

	return nil
}

// validateDatabaseConfig validates the database configuration.
func (c *Config) validateDatabaseConfig() error {
	if c.LocalDB {
		return nil // Skip validation for local DB
	}

	if c.DBHost == "" {
		return errors.New("database host is required when not using local DB")
	}

	if c.DBUser == "" {
		return errors.New("database user is required when not using local DB")
	}

	if c.DBPassword == "" {
		return errors.New("database password is required when not using local DB")
	}

	if c.DBName == "" {
		return errors.New("database name is required when not using local DB")
	}

	if c.DBPort <= 0 || c.DBPort > MaxPortNumber {
		return errors.New("database port must be between 1 and 65535")
	}

	return nil
}

// validateConnectionPool validates the connection pool settings.
func (c *Config) validateConnectionPool() error {
	if c.DBMaxIdleConns < 0 {
		return errors.New("database max idle connections cannot be negative")
	}

	if c.DBMaxOpenConns < 0 {
		return errors.New("database max open connections cannot be negative")
	}

	if c.DBMaxIdleConns > c.DBMaxOpenConns && c.DBMaxOpenConns > 0 {
		return errors.New("database max idle connections cannot exceed max open connections")
	}

	return nil
}

// validateLogLevel validates the log level configuration.
func (c *Config) validateLogLevel() error {
	validLogLevels := []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}

	for _, level := range validLogLevels {
		if strings.EqualFold(c.LogLevel, level) {
			return nil
		}
	}

	return fmt.Errorf("log level must be one of: %v", validLogLevels)
}

// validateTLSConfig validates the TLS configuration.
func (c *Config) validateTLSConfig() error {
	if !c.TLSEnabled {
		return nil
	}

	if c.TLSCertFile == "" {
		return errors.New("TLS certificate file is required when TLS is enabled")
	}

	if c.TLSKeyFile == "" {
		return errors.New("TLS key file is required when TLS is enabled")
	}

	return nil
}

// validateAuthConfig validates the authentication configuration.
func (c *Config) validateAuthConfig() error {
	if !c.RequireAuth {
		return nil
	}

	if c.AuthUsername == "" {
		return errors.New("auth username is required when authentication is enabled")
	}

	if c.AuthPassword == "" {
		return errors.New("auth password is required when authentication is enabled")
	}

	return nil
}

// validateRequestSize validates the request size limits.
func (c *Config) validateRequestSize() error {
	if c.MaxRequestSize < MinRequestSize {
		return errors.New("max request size must be at least 1MB")
	}

	if c.MaxRequestSize > MaxRequestSize {
		return errors.New("max request size cannot exceed 100MB")
	}

	return nil
}

// ObservabilityConfig holds observability configuration values.
type ObservabilityConfig struct {
	TracingEnabled  bool
	MetricsEnabled  bool
	ServiceName     string
	Version         string
	Environment     string
	TraceEndpoint   string
	MetricsEndpoint string
	Interval        time.Duration
}

// GetObservabilityConfig returns observability configuration values.
func (c *Config) GetObservabilityConfig() ObservabilityConfig {
	return ObservabilityConfig{
		TracingEnabled:  c.TracingEnabled,
		MetricsEnabled:  c.MetricsEnabled,
		ServiceName:     c.ServiceName,
		Version:         c.ServiceVersion,
		Environment:     c.Environment,
		TraceEndpoint:   c.OTLPTraceEndpoint,
		MetricsEndpoint: c.OTLPMetricsEndpoint,
		Interval:        time.Duration(c.MetricsInterval) * time.Second,
	}
}
