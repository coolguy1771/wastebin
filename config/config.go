package config

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

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
	AllowedOrigins    string `koanf:"ALLOWED_ORIGINS"`
	RequireAuth       bool   `koanf:"REQUIRE_AUTH"`
	AuthUsername      string `koanf:"AUTH_USERNAME"`
	AuthPassword      string `koanf:"AUTH_PASSWORD"`
	CSRFKey           string `koanf:"CSRF_KEY"`
	MaxRequestSize    int64  `koanf:"MAX_REQUEST_SIZE"`

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

// Load initializes the configuration by loading defaults and environment variables.
func Load() *Config {
	k := koanf.New(".")

	// Load default configuration settings
	k.Load(confmap.Provider(map[string]interface{}{
		"WEBAPP_PORT":           "3000",
		"DB_MAX_IDLE_CONNS":     10,
		"DB_MAX_OPEN_CONNS":     50,
		"DB_PORT":               5432,
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
		"MAX_REQUEST_SIZE":      15728640, // 15MB
		"TRACING_ENABLED":       true,
		"METRICS_ENABLED":       true,
		"SERVICE_NAME":          "wastebin",
		"SERVICE_VERSION":       "1.0.0",
		"ENVIRONMENT":           "development",
		"OTLP_TRACE_ENDPOINT":   "localhost:4318",
		"OTLP_METRICS_ENDPOINT": "localhost:4318",
		"METRICS_INTERVAL":      30,
	}, "."), nil)

	// Load environment variables with prefix WASTEBIN_
	k.Load(env.Provider("WASTEBIN_", ".", func(s string) string {
		return strings.TrimPrefix(s, "WASTEBIN_")
	}), nil)

	// Unmarshal the configuration into the Config struct
	if err := k.Unmarshal("", &Conf); err != nil {
		log.Fatal("Error loading config", zap.Error(err))
	}

	// Validate the configuration
	if err := Conf.Validate(); err != nil {
		log.Fatal("Configuration validation failed", zap.Error(err))
	}

	return &Conf
}

// Validate validates the configuration settings
func (c *Config) Validate() error {
	var errs []string

	// Validate webapp port
	if c.WebappPort == "" {
		errs = append(errs, "webapp port must be specified")
	} else {
		if port, err := strconv.Atoi(c.WebappPort); err != nil || port < 1 || port > 65535 {
			errs = append(errs, "webapp port must be a valid port number (1-65535)")
		}
	}

	// Validate database configuration for PostgreSQL
	if !c.LocalDB {
		if c.DBHost == "" {
			errs = append(errs, "database host is required when not using local DB")
		}
		if c.DBUser == "" {
			errs = append(errs, "database user is required when not using local DB")
		}
		if c.DBPassword == "" {
			errs = append(errs, "database password is required when not using local DB")
		}
		if c.DBName == "" {
			errs = append(errs, "database name is required when not using local DB")
		}
		if c.DBPort <= 0 || c.DBPort > 65535 {
			errs = append(errs, "database port must be between 1 and 65535")
		}
	}

	// Validate connection pool settings
	if c.DBMaxIdleConns < 0 {
		errs = append(errs, "database max idle connections cannot be negative")
	}
	if c.DBMaxOpenConns < 0 {
		errs = append(errs, "database max open connections cannot be negative")
	}
	if c.DBMaxIdleConns > c.DBMaxOpenConns && c.DBMaxOpenConns > 0 {
		errs = append(errs, "database max idle connections cannot exceed max open connections")
	}

	// Validate log level
	validLogLevels := []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}
	validLevel := false
	for _, level := range validLogLevels {
		if strings.ToUpper(c.LogLevel) == level {
			validLevel = true
			break
		}
	}
	if !validLevel {
		errs = append(errs, fmt.Sprintf("log level must be one of: %v", validLogLevels))
	}

	// Validate TLS configuration
	if c.TLSEnabled {
		if c.TLSCertFile == "" {
			errs = append(errs, "TLS certificate file is required when TLS is enabled")
		}
		if c.TLSKeyFile == "" {
			errs = append(errs, "TLS key file is required when TLS is enabled")
		}
	}

	// Validate authentication configuration
	if c.RequireAuth {
		if c.AuthUsername == "" {
			errs = append(errs, "auth username is required when authentication is enabled")
		}
		if c.AuthPassword == "" {
			errs = append(errs, "auth password is required when authentication is enabled")
		}
	}

	// Validate request size limits
	if c.MaxRequestSize < 1024*1024 { // Minimum 1MB
		errs = append(errs, "max request size must be at least 1MB")
	}
	if c.MaxRequestSize > 100*1024*1024 { // Maximum 100MB
		errs = append(errs, "max request size cannot exceed 100MB")
	}

	if len(errs) > 0 {
		return errors.New("configuration validation errors: " + strings.Join(errs, "; "))
	}

	return nil
}

// GetObservabilityConfig returns observability configuration values
func (c *Config) GetObservabilityConfig() (tracingEnabled, metricsEnabled bool, serviceName, version, environment, traceEndpoint, metricsEndpoint string, interval time.Duration) {
	return c.TracingEnabled,
		c.MetricsEnabled,
		c.ServiceName,
		c.ServiceVersion,
		c.Environment,
		c.OTLPTraceEndpoint,
		c.OTLPMetricsEndpoint,
		time.Duration(c.MetricsInterval) * time.Second
}
