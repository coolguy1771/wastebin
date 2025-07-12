package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDefaultConfig(t *testing.T) {
	// Clear any existing environment variables
	clearWastebinEnvVars()

	// Set minimal required config for validation to pass
	os.Setenv("WASTEBIN_LOCAL_DB", "true")
	defer os.Unsetenv("WASTEBIN_LOCAL_DB")

	config := Load()

	// Test default values
	assert.Equal(t, "3000", config.WebappPort)
	assert.Equal(t, 10, config.DBMaxIdleConns)
	assert.Equal(t, 50, config.DBMaxOpenConns)
	assert.Equal(t, 5432, config.DBPort)
	assert.Equal(t, "localhost", config.DBHost)
	assert.Equal(t, "wastebin", config.DBUser)
	assert.Equal(t, "wastebin", config.DBName)
	assert.Equal(t, "INFO", config.LogLevel)
	assert.Equal(t, true, config.LocalDB)
	assert.Equal(t, false, config.Dev)
}

func TestLoadEnvironmentVariables(t *testing.T) {
	// Clear any existing environment variables
	clearWastebinEnvVars()

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
		os.Setenv(key, value)
		defer os.Unsetenv(key)
	}

	config := Load()

	// Test environment variable overrides
	assert.Equal(t, "8080", config.WebappPort)
	assert.Equal(t, 15, config.DBMaxIdleConns)
	assert.Equal(t, 75, config.DBMaxOpenConns)
	assert.Equal(t, 5433, config.DBPort)
	assert.Equal(t, "testhost", config.DBHost)
	assert.Equal(t, "testuser", config.DBUser)
	assert.Equal(t, "testdb", config.DBName)
	assert.Equal(t, "testpass", config.DBPassword)
	assert.Equal(t, "DEBUG", config.LogLevel)
	assert.Equal(t, false, config.LocalDB)
	assert.Equal(t, true, config.Dev)
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name          string
		config        Config
		expectError   bool
		errorContains string
	}{
		{
			name: "Valid local DB config",
			config: Config{
				WebappPort:     "3000",
				DBMaxIdleConns: 5,
				DBMaxOpenConns: 10,
				LogLevel:       "INFO",
				LocalDB:        true,
			},
			expectError: false,
		},
		{
			name: "Valid PostgreSQL config",
			config: Config{
				WebappPort:     "3000",
				DBHost:         "localhost",
				DBUser:         "user",
				DBPassword:     "password",
				DBName:         "database",
				DBPort:         5432,
				DBMaxIdleConns: 5,
				DBMaxOpenConns: 10,
				LogLevel:       "INFO",
				LocalDB:        false,
			},
			expectError: false,
		},
		{
			name: "Empty webapp port",
			config: Config{
				WebappPort: "",
				LocalDB:    true,
				LogLevel:   "INFO",
			},
			expectError:   true,
			errorContains: "webapp port must be specified",
		},
		{
			name: "Invalid webapp port",
			config: Config{
				WebappPort: "invalid",
				LocalDB:    true,
				LogLevel:   "INFO",
			},
			expectError:   true,
			errorContains: "webapp port must be a valid port number",
		},
		{
			name: "Port out of range",
			config: Config{
				WebappPort: "99999",
				LocalDB:    true,
				LogLevel:   "INFO",
			},
			expectError:   true,
			errorContains: "webapp port must be a valid port number",
		},
		{
			name: "Missing DB host for PostgreSQL",
			config: Config{
				WebappPort: "3000",
				DBUser:     "user",
				DBPassword: "password",
				DBName:     "database",
				DBPort:     5432,
				LogLevel:   "INFO",
				LocalDB:    false,
			},
			expectError:   true,
			errorContains: "database host is required",
		},
		{
			name: "Missing DB user for PostgreSQL",
			config: Config{
				WebappPort: "3000",
				DBHost:     "localhost",
				DBPassword: "password",
				DBName:     "database",
				DBPort:     5432,
				LogLevel:   "INFO",
				LocalDB:    false,
			},
			expectError:   true,
			errorContains: "database user is required",
		},
		{
			name: "Missing DB password for PostgreSQL",
			config: Config{
				WebappPort: "3000",
				DBHost:     "localhost",
				DBUser:     "user",
				DBName:     "database",
				DBPort:     5432,
				LogLevel:   "INFO",
				LocalDB:    false,
			},
			expectError:   true,
			errorContains: "database password is required",
		},
		{
			name: "Missing DB name for PostgreSQL",
			config: Config{
				WebappPort: "3000",
				DBHost:     "localhost",
				DBUser:     "user",
				DBPassword: "password",
				DBPort:     5432,
				LogLevel:   "INFO",
				LocalDB:    false,
			},
			expectError:   true,
			errorContains: "database name is required",
		},
		{
			name: "Invalid DB port",
			config: Config{
				WebappPort: "3000",
				DBHost:     "localhost",
				DBUser:     "user",
				DBPassword: "password",
				DBName:     "database",
				DBPort:     99999,
				LogLevel:   "INFO",
				LocalDB:    false,
			},
			expectError:   true,
			errorContains: "database port must be between 1 and 65535",
		},
		{
			name: "Negative max idle connections",
			config: Config{
				WebappPort:     "3000",
				DBMaxIdleConns: -1,
				DBMaxOpenConns: 10,
				LogLevel:       "INFO",
				LocalDB:        true,
			},
			expectError:   true,
			errorContains: "database max idle connections cannot be negative",
		},
		{
			name: "Negative max open connections",
			config: Config{
				WebappPort:     "3000",
				DBMaxIdleConns: 5,
				DBMaxOpenConns: -1,
				LogLevel:       "INFO",
				LocalDB:        true,
			},
			expectError:   true,
			errorContains: "database max open connections cannot be negative",
		},
		{
			name: "Idle connections exceed open connections",
			config: Config{
				WebappPort:     "3000",
				DBMaxIdleConns: 20,
				DBMaxOpenConns: 10,
				LogLevel:       "INFO",
				LocalDB:        true,
			},
			expectError:   true,
			errorContains: "database max idle connections cannot exceed max open connections",
		},
		{
			name: "Invalid log level",
			config: Config{
				WebappPort: "3000",
				LogLevel:   "INVALID",
				LocalDB:    true,
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
	config := Config{
		WebappPort:     "3000",
		DBMaxIdleConns: 0, // Should be allowed (will use defaults)
		DBMaxOpenConns: 0, // Should be allowed (will use defaults)
		LogLevel:       "INFO",
		LocalDB:        true,
	}

	err := config.Validate()
	assert.NoError(t, err)
}

func TestLogLevelCaseInsensitive(t *testing.T) {
	logLevels := []string{"debug", "info", "warn", "error", "fatal", "DEBUG", "INFO", "WARN", "ERROR", "FATAL"}

	for _, level := range logLevels {
		t.Run("LogLevel-"+level, func(t *testing.T) {
			config := Config{
				WebappPort: "3000",
				LogLevel:   level,
				LocalDB:    true,
			}

			err := config.Validate()
			assert.NoError(t, err, "Log level %s should be valid", level)
		})
	}
}

// Helper function to clear all WASTEBIN_ environment variables
func clearWastebinEnvVars() {
	envVars := []string{
		"WASTEBIN_WEBAPP_PORT",
		"WASTEBIN_DB_MAX_IDLE_CONNS",
		"WASTEBIN_DB_MAX_OPEN_CONNS",
		"WASTEBIN_DB_PORT",
		"WASTEBIN_DB_HOST",
		"WASTEBIN_DB_USER",
		"WASTEBIN_DB_NAME",
		"WASTEBIN_DB_PASSWORD",
		"WASTEBIN_LOG_LEVEL",
		"WASTEBIN_LOCAL_DB",
		"WASTEBIN_DEV",
	}

	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}
}

func BenchmarkLoadConfig(b *testing.B) {
	// Set minimal required config
	os.Setenv("WASTEBIN_LOCAL_DB", "true")
	defer os.Unsetenv("WASTEBIN_LOCAL_DB")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Load()
	}
}

func BenchmarkValidateConfig(b *testing.B) {
	config := Config{
		WebappPort:     "3000",
		DBMaxIdleConns: 5,
		DBMaxOpenConns: 10,
		LogLevel:       "INFO",
		LocalDB:        true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.Validate()
	}
}
