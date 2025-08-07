package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coolguy1771/wastebin/config"
)

func TestLoadDefaultConfig(t *testing.T) {
	// Clear any existing environment variables
	clearWastebinEnvVars(t)

	// Set minimal required config for validation to pass
	t.Setenv("WASTEBIN_LOCAL_DB", "true")
	defer os.Unsetenv("WASTEBIN_LOCAL_DB")

	cfg := config.Load()

	// Test default values
	assert.Equal(t, "3000", cfg.WebappPort)
	assert.Equal(t, 10, cfg.DBMaxIdleConns)
	assert.Equal(t, 50, cfg.DBMaxOpenConns)
	assert.Equal(t, 5432, cfg.DBPort)
	assert.Equal(t, "localhost", cfg.DBHost)
	assert.Equal(t, "wastebin", cfg.DBUser)
	assert.Equal(t, "wastebin", cfg.DBName)
	assert.Equal(t, "INFO", cfg.LogLevel)
	assert.True(t, cfg.LocalDB)
	assert.False(t, cfg.Dev)
}

func TestLoadEnvironmentVariables(t *testing.T) {
	// Clear any existing environment variables
	clearWastebinEnvVars(t)

	// Set environment variables
	envVars := map[string]string{
		"WASTEBIN_WEBAPP_PORT":       "8080",
		"WASTEBIN_DB_MAX_IDLE_CONNS": "15",
		"WASTEBIN_DB_MAX_OPEN_CONNS": "75",
		"WASTEBIN_DB_PORT":           "5433",
		"WASTEBIN_DB_HOST":           "testhost",
		"WASTEBIN_DB_USER":           "testuser",
		"WASTEBIN_DB_NAME":           "testdb",
		"WASTEBIN_DB_PASSWORD":       "testpass",
		"WASTEBIN_LOG_LEVEL":         "DEBUG",
		"WASTEBIN_LOCAL_DB":          "false",
		"WASTEBIN_DEV":               "true",
	}

	for key, value := range envVars {
		t.Setenv(key, value)
		defer os.Unsetenv(key)
	}

	cfg := config.Load()

	// Test environment variable overrides
	assert.Equal(t, "8080", cfg.WebappPort)
	assert.Equal(t, 15, cfg.DBMaxIdleConns)
	assert.Equal(t, 75, cfg.DBMaxOpenConns)
	assert.Equal(t, 5433, cfg.DBPort)
	assert.Equal(t, "testhost", cfg.DBHost)
	assert.Equal(t, "testuser", cfg.DBUser)
	assert.Equal(t, "testdb", cfg.DBName)
	assert.Equal(t, "testpass", cfg.DBPassword)
	assert.Equal(t, "DEBUG", cfg.LogLevel)
	assert.False(t, cfg.LocalDB)
	assert.True(t, cfg.Dev)
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name          string
		config        config.Config
		expectError   bool
		errorContains string
	}{
		{
			name: "Valid local DB config",
			config: config.Config{
				WebappPort:          "3000",
				DBMaxIdleConns:      5,
				DBMaxOpenConns:      10,
				LogLevel:            "INFO",
				LocalDB:             true,
				DBUser:              "",
				DBPassword:          "",
				DBHost:              "",
				DBPort:              0,
				Dev:                 false,
				TLSEnabled:          false,
				TLSCertFile:         "",
				TLSKeyFile:          "",
				AllowedOrigins:      "",
				RequireAuth:         false,
				AuthUsername:        "",
				AuthPassword:        "",
				CSRFKey:             "",
				MaxRequestSize:      1024 * 1024, // 1MB minimum
				TracingEnabled:      false,
				MetricsEnabled:      false,
				ServiceName:         "",
				ServiceVersion:      "",
				Environment:         "",
				OTLPTraceEndpoint:   "",
				OTLPMetricsEndpoint: "",
				MetricsInterval:     0,
			},
			expectError: false,
		},
		{
			name: "Valid PostgreSQL config",
			config: config.Config{
				WebappPort:          "3000",
				DBHost:              "localhost",
				DBUser:              "user",
				DBPassword:          "password",
				DBName:              "database",
				DBPort:              5432,
				DBMaxIdleConns:      5,
				DBMaxOpenConns:      10,
				LogLevel:            "INFO",
				LocalDB:             false,
				Dev:                 false,
				TLSEnabled:          false,
				TLSCertFile:         "",
				TLSKeyFile:          "",
				AllowedOrigins:      "",
				RequireAuth:         false,
				AuthUsername:        "",
				AuthPassword:        "",
				CSRFKey:             "",
				MaxRequestSize:      1024 * 1024, // 1MB minimum
				TracingEnabled:      false,
				MetricsEnabled:      false,
				ServiceName:         "",
				ServiceVersion:      "",
				Environment:         "",
				OTLPTraceEndpoint:   "",
				OTLPMetricsEndpoint: "",
				MetricsInterval:     0,
			},
			expectError: false,
		},
		{
			name: "Empty webapp port",
			config: config.Config{
				WebappPort:          "",
				LocalDB:             true,
				LogLevel:            "INFO",
				DBUser:              "",
				DBPassword:          "",
				DBHost:              "",
				DBPort:              0,
				DBMaxIdleConns:      0,
				DBMaxOpenConns:      0,
				Dev:                 false,
				TLSEnabled:          false,
				TLSCertFile:         "",
				TLSKeyFile:          "",
				AllowedOrigins:      "",
				RequireAuth:         false,
				AuthUsername:        "",
				AuthPassword:        "",
				CSRFKey:             "",
				MaxRequestSize:      1024 * 1024, // 1MB minimum
				TracingEnabled:      false,
				MetricsEnabled:      false,
				ServiceName:         "",
				ServiceVersion:      "",
				Environment:         "",
				OTLPTraceEndpoint:   "",
				OTLPMetricsEndpoint: "",
				MetricsInterval:     0,
			},
			expectError:   true,
			errorContains: "webapp port must be specified",
		},
		{
			name: "Invalid webapp port",
			config: config.Config{
				WebappPort:          "invalid",
				LocalDB:             true,
				LogLevel:            "INFO",
				DBUser:              "",
				DBPassword:          "",
				DBHost:              "",
				DBPort:              0,
				DBMaxIdleConns:      0,
				DBMaxOpenConns:      0,
				Dev:                 false,
				TLSEnabled:          false,
				TLSCertFile:         "",
				TLSKeyFile:          "",
				AllowedOrigins:      "",
				RequireAuth:         false,
				AuthUsername:        "",
				AuthPassword:        "",
				CSRFKey:             "",
				MaxRequestSize:      1024 * 1024, // 1MB minimum
				TracingEnabled:      false,
				MetricsEnabled:      false,
				ServiceName:         "",
				ServiceVersion:      "",
				Environment:         "",
				OTLPTraceEndpoint:   "",
				OTLPMetricsEndpoint: "",
				MetricsInterval:     0,
			},
			expectError:   true,
			errorContains: "webapp port must be a valid port number",
		},
		{
			name: "Port out of range",
			config: config.Config{
				WebappPort:          "99999",
				LocalDB:             true,
				LogLevel:            "INFO",
				DBUser:              "",
				DBPassword:          "",
				DBHost:              "",
				DBPort:              0,
				DBMaxIdleConns:      0,
				DBMaxOpenConns:      0,
				Dev:                 false,
				TLSEnabled:          false,
				TLSCertFile:         "",
				TLSKeyFile:          "",
				AllowedOrigins:      "",
				RequireAuth:         false,
				AuthUsername:        "",
				AuthPassword:        "",
				CSRFKey:             "",
				MaxRequestSize:      1024 * 1024, // 1MB minimum
				TracingEnabled:      false,
				MetricsEnabled:      false,
				ServiceName:         "",
				ServiceVersion:      "",
				Environment:         "",
				OTLPTraceEndpoint:   "",
				OTLPMetricsEndpoint: "",
				MetricsInterval:     0,
			},
			expectError:   true,
			errorContains: "webapp port must be a valid port number",
		},
		{
			name: "Missing DB host for PostgreSQL",
			config: config.Config{
				WebappPort:          "3000",
				DBUser:              "user",
				DBPassword:          "password",
				DBName:              "database",
				DBPort:              5432,
				LogLevel:            "INFO",
				LocalDB:             false,
				DBMaxIdleConns:      0,
				DBMaxOpenConns:      0,
				Dev:                 false,
				TLSEnabled:          false,
				TLSCertFile:         "",
				TLSKeyFile:          "",
				AllowedOrigins:      "",
				RequireAuth:         false,
				AuthUsername:        "",
				AuthPassword:        "",
				CSRFKey:             "",
				MaxRequestSize:      1024 * 1024, // 1MB minimum
				TracingEnabled:      false,
				MetricsEnabled:      false,
				ServiceName:         "",
				ServiceVersion:      "",
				Environment:         "",
				OTLPTraceEndpoint:   "",
				OTLPMetricsEndpoint: "",
				MetricsInterval:     0,
			},
			expectError:   true,
			errorContains: "database host is required",
		},
		{
			name: "Missing DB user for PostgreSQL",
			config: config.Config{
				WebappPort:          "3000",
				DBHost:              "localhost",
				DBPassword:          "password",
				DBName:              "database",
				DBPort:              5432,
				LogLevel:            "INFO",
				LocalDB:             false,
				DBUser:              "",
				DBMaxIdleConns:      0,
				DBMaxOpenConns:      0,
				Dev:                 false,
				TLSEnabled:          false,
				TLSCertFile:         "",
				TLSKeyFile:          "",
				AllowedOrigins:      "",
				RequireAuth:         false,
				AuthUsername:        "",
				AuthPassword:        "",
				CSRFKey:             "",
				MaxRequestSize:      1024 * 1024, // 1MB minimum
				TracingEnabled:      false,
				MetricsEnabled:      false,
				ServiceName:         "",
				ServiceVersion:      "",
				Environment:         "",
				OTLPTraceEndpoint:   "",
				OTLPMetricsEndpoint: "",
				MetricsInterval:     0,
			},
			expectError:   true,
			errorContains: "database user is required",
		},
		{
			name: "Missing DB password for PostgreSQL",
			config: config.Config{
				WebappPort:          "3000",
				DBHost:              "localhost",
				DBUser:              "user",
				DBName:              "database",
				DBPort:              5432,
				LogLevel:            "INFO",
				LocalDB:             false,
				DBPassword:          "",
				DBMaxIdleConns:      0,
				DBMaxOpenConns:      0,
				Dev:                 false,
				TLSEnabled:          false,
				TLSCertFile:         "",
				TLSKeyFile:          "",
				AllowedOrigins:      "",
				RequireAuth:         false,
				AuthUsername:        "",
				AuthPassword:        "",
				CSRFKey:             "",
				MaxRequestSize:      1024 * 1024, // 1MB minimum
				TracingEnabled:      false,
				MetricsEnabled:      false,
				ServiceName:         "",
				ServiceVersion:      "",
				Environment:         "",
				OTLPTraceEndpoint:   "",
				OTLPMetricsEndpoint: "",
				MetricsInterval:     0,
			},
			expectError:   true,
			errorContains: "database password is required",
		},
		{
			name: "Missing DB name for PostgreSQL",
			config: config.Config{
				WebappPort:          "3000",
				DBHost:              "localhost",
				DBUser:              "user",
				DBPassword:          "password",
				DBPort:              5432,
				LogLevel:            "INFO",
				LocalDB:             false,
				DBName:              "",
				DBMaxIdleConns:      0,
				DBMaxOpenConns:      0,
				Dev:                 false,
				TLSEnabled:          false,
				TLSCertFile:         "",
				TLSKeyFile:          "",
				AllowedOrigins:      "",
				RequireAuth:         false,
				AuthUsername:        "",
				AuthPassword:        "",
				CSRFKey:             "",
				MaxRequestSize:      1024 * 1024, // 1MB minimum
				TracingEnabled:      false,
				MetricsEnabled:      false,
				ServiceName:         "",
				ServiceVersion:      "",
				Environment:         "",
				OTLPTraceEndpoint:   "",
				OTLPMetricsEndpoint: "",
				MetricsInterval:     0,
			},
			expectError:   true,
			errorContains: "database name is required",
		},
		{
			name: "Invalid DB port",
			config: config.Config{
				WebappPort:          "3000",
				DBHost:              "localhost",
				DBUser:              "user",
				DBPassword:          "password",
				DBName:              "database",
				LogLevel:            "INFO",
				LocalDB:             false,
				DBPort:              99999,
				DBMaxIdleConns:      0,
				DBMaxOpenConns:      0,
				Dev:                 false,
				TLSEnabled:          false,
				TLSCertFile:         "",
				TLSKeyFile:          "",
				AllowedOrigins:      "",
				RequireAuth:         false,
				AuthUsername:        "",
				AuthPassword:        "",
				CSRFKey:             "",
				MaxRequestSize:      1024 * 1024, // 1MB minimum
				TracingEnabled:      false,
				MetricsEnabled:      false,
				ServiceName:         "",
				ServiceVersion:      "",
				Environment:         "",
				OTLPTraceEndpoint:   "",
				OTLPMetricsEndpoint: "",
				MetricsInterval:     0,
			},
			expectError:   true,
			errorContains: "database port must be between 1 and 65535",
		},
		{
			name: "Negative max idle connections",
			config: config.Config{
				WebappPort:          "3000",
				DBMaxIdleConns:      -1,
				DBMaxOpenConns:      10,
				LogLevel:            "INFO",
				LocalDB:             true,
				DBUser:              "",
				DBPassword:          "",
				DBHost:              "",
				DBPort:              0,
				Dev:                 false,
				TLSEnabled:          false,
				TLSCertFile:         "",
				TLSKeyFile:          "",
				AllowedOrigins:      "",
				RequireAuth:         false,
				AuthUsername:        "",
				AuthPassword:        "",
				CSRFKey:             "",
				MaxRequestSize:      1024 * 1024, // 1MB minimum
				TracingEnabled:      false,
				MetricsEnabled:      false,
				ServiceName:         "",
				ServiceVersion:      "",
				Environment:         "",
				OTLPTraceEndpoint:   "",
				OTLPMetricsEndpoint: "",
				MetricsInterval:     0,
			},
			expectError:   true,
			errorContains: "database max idle connections cannot be negative",
		},
		{
			name: "Negative max open connections",
			config: config.Config{
				WebappPort:          "3000",
				DBMaxIdleConns:      5,
				DBMaxOpenConns:      -1,
				LogLevel:            "INFO",
				LocalDB:             true,
				DBUser:              "",
				DBPassword:          "",
				DBHost:              "",
				DBPort:              0,
				Dev:                 false,
				TLSEnabled:          false,
				TLSCertFile:         "",
				TLSKeyFile:          "",
				AllowedOrigins:      "",
				RequireAuth:         false,
				AuthUsername:        "",
				AuthPassword:        "",
				CSRFKey:             "",
				MaxRequestSize:      1024 * 1024, // 1MB minimum
				TracingEnabled:      false,
				MetricsEnabled:      false,
				ServiceName:         "",
				ServiceVersion:      "",
				Environment:         "",
				OTLPTraceEndpoint:   "",
				OTLPMetricsEndpoint: "",
				MetricsInterval:     0,
			},
			expectError:   true,
			errorContains: "database max open connections cannot be negative",
		},
		{
			name: "Idle connections exceed open connections",
			config: config.Config{
				WebappPort:          "3000",
				DBMaxIdleConns:      20,
				DBMaxOpenConns:      10,
				LogLevel:            "INFO",
				LocalDB:             true,
				DBUser:              "",
				DBPassword:          "",
				DBHost:              "",
				DBPort:              0,
				Dev:                 false,
				TLSEnabled:          false,
				TLSCertFile:         "",
				TLSKeyFile:          "",
				AllowedOrigins:      "",
				RequireAuth:         false,
				AuthUsername:        "",
				AuthPassword:        "",
				CSRFKey:             "",
				MaxRequestSize:      1024 * 1024, // 1MB minimum
				TracingEnabled:      false,
				MetricsEnabled:      false,
				ServiceName:         "",
				ServiceVersion:      "",
				Environment:         "",
				OTLPTraceEndpoint:   "",
				OTLPMetricsEndpoint: "",
				MetricsInterval:     0,
			},
			expectError:   true,
			errorContains: "database max idle connections cannot exceed max open connections",
		},
		{
			name: "Invalid log level",
			config: config.Config{
				WebappPort:          "3000",
				LogLevel:            "INVALID",
				LocalDB:             true,
				DBUser:              "",
				DBPassword:          "",
				DBHost:              "",
				DBPort:              0,
				DBMaxIdleConns:      0,
				DBMaxOpenConns:      0,
				Dev:                 false,
				TLSEnabled:          false,
				TLSCertFile:         "",
				TLSKeyFile:          "",
				AllowedOrigins:      "",
				RequireAuth:         false,
				AuthUsername:        "",
				AuthPassword:        "",
				CSRFKey:             "",
				MaxRequestSize:      1024 * 1024, // 1MB minimum
				TracingEnabled:      false,
				MetricsEnabled:      false,
				ServiceName:         "",
				ServiceVersion:      "",
				Environment:         "",
				OTLPTraceEndpoint:   "",
				OTLPMetricsEndpoint: "",
				MetricsInterval:     0,
			},
			expectError:   true,
			errorContains: "log level must be one of",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				require.Error(t, err)

				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigValidationWithZeroValues(t *testing.T) {
	// Test that zero values for connection pool settings are allowed
	cfg := config.Config{
		WebappPort:          "3000",
		DBMaxIdleConns:      0, // Should be allowed (will use defaults)
		DBMaxOpenConns:      0, // Should be allowed (will use defaults)
		LogLevel:            "INFO",
		LocalDB:             true,
		DBUser:              "",
		DBPassword:          "",
		DBHost:              "",
		DBPort:              0,
		Dev:                 false,
		TLSEnabled:          false,
		TLSCertFile:         "",
		TLSKeyFile:          "",
		AllowedOrigins:      "",
		RequireAuth:         false,
		AuthUsername:        "",
		AuthPassword:        "",
		CSRFKey:             "",
		MaxRequestSize:      1024 * 1024, // 1MB minimum
		TracingEnabled:      false,
		MetricsEnabled:      false,
		ServiceName:         "",
		ServiceVersion:      "",
		Environment:         "",
		OTLPTraceEndpoint:   "",
		OTLPMetricsEndpoint: "",
		MetricsInterval:     0,
	}

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestLogLevelCaseInsensitive(t *testing.T) {
	logLevels := []string{"debug", "INFO", "Warn", "error", "FATAL"}
	for _, level := range logLevels {
		t.Run(level, func(t *testing.T) {
			cfg := config.Config{
				WebappPort:          "3000",
				DBMaxIdleConns:      5,
				DBMaxOpenConns:      10,
				LogLevel:            level,
				LocalDB:             true,
				DBUser:              "",
				DBPassword:          "",
				DBHost:              "",
				DBPort:              0,
				Dev:                 false,
				TLSEnabled:          false,
				TLSCertFile:         "",
				TLSKeyFile:          "",
				AllowedOrigins:      "",
				RequireAuth:         false,
				AuthUsername:        "",
				AuthPassword:        "",
				CSRFKey:             "",
				MaxRequestSize:      1024 * 1024, // 1MB minimum
				TracingEnabled:      false,
				MetricsEnabled:      false,
				ServiceName:         "",
				ServiceVersion:      "",
				Environment:         "",
				OTLPTraceEndpoint:   "",
				OTLPMetricsEndpoint: "",
				MetricsInterval:     0,
			}

			err := cfg.Validate()
			assert.NoError(t, err, "Log level %s should be valid", level)
		})
	}
}

// clearWastebinEnvVars unsets all WASTEBIN_ environment variables for test isolation.
func clearWastebinEnvVars(t *testing.T) {
	t.Helper()

	for _, env := range os.Environ() {
		if len(env) > 9 && env[:9] == "WASTEBIN_" {
			parts := []rune(env)
			for i, c := range parts {
				if c == '=' {
					os.Unsetenv(string(parts[:i]))

					break
				}
			}
		}
	}
}

func BenchmarkLoadConfig(b *testing.B) {
	// Set minimal required config
	b.Setenv("WASTEBIN_LOCAL_DB", "true")
	defer os.Unsetenv("WASTEBIN_LOCAL_DB")

	b.ResetTimer()

	for range b.N {
		config.Load()
	}
}

func BenchmarkValidateConfig(b *testing.B) {
	config := config.Config{
		WebappPort:          "3000",
		DBMaxIdleConns:      5,
		DBMaxOpenConns:      10,
		LogLevel:            "INFO",
		LocalDB:             true,
		DBUser:              "",
		DBPassword:          "",
		DBHost:              "",
		DBPort:              0,
		Dev:                 false,
		TLSEnabled:          false,
		TLSCertFile:         "",
		TLSKeyFile:          "",
		AllowedOrigins:      "",
		RequireAuth:         false,
		AuthUsername:        "",
		AuthPassword:        "",
		CSRFKey:             "",
		MaxRequestSize:      1024 * 1024, // 1MB minimum
		TracingEnabled:      false,
		MetricsEnabled:      false,
		ServiceName:         "",
		ServiceVersion:      "",
		Environment:         "",
		OTLPTraceEndpoint:   "",
		OTLPMetricsEndpoint: "",
		MetricsInterval:     0,
	}

	b.ResetTimer()

	for range b.N {
		config.Validate()
	}
}
