//go:build integration
// +build integration

package storage_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/coolguy1771/wastebin/config"
	"github.com/coolguy1771/wastebin/log"
	"github.com/coolguy1771/wastebin/models"
	"github.com/coolguy1771/wastebin/storage"
)

func TestMain(m *testing.M) {
	// Initialize logger for tests
	logger, err := log.New(os.Stdout, "ERROR")
	if err != nil {
		panic(err)
	}
	log.ResetDefault(logger)
	
	// Run tests
	code := m.Run()
	os.Exit(code)
}

// setupTestConfig initializes the configuration for testing.
func setupTestConfig(t *testing.T, localDB bool) {
	t.Helper()

	//nolint:reassign // Conf is reassigned for test setup.
	config.Conf = config.Config{
		LocalDB:             localDB,
		DBUser:              "test_user",
		DBPassword:          "test_pass",
		DBHost:              "localhost",
		DBPort:              5432,
		DBName:              "test_db",
		DBMaxIdleConns:      5,
		DBMaxOpenConns:      10,
		WebappPort:          "",
		Dev:                 false,
		TLSEnabled:          false,
		TLSCertFile:         "",
		TLSKeyFile:          "",
		AllowedOrigins:      "",
		RequireAuth:         false,
		AuthUsername:        "",
		AuthPassword:        "",
		CSRFKey:             "",
		MaxRequestSize:      0,
		LogLevel:            "",
		TracingEnabled:      false,
		MetricsEnabled:      false,
		ServiceName:         "",
		ServiceVersion:      "",
		Environment:         "",
		OTLPTraceEndpoint:   "",
		OTLPMetricsEndpoint: "",
		MetricsInterval:     0,
	}
}

func TestConnectSQLite(t *testing.T) {
	setupTestConfig(t, true)

	err := storage.Connect(nil)
	require.NoError(t, err, "Expected no error when connecting to SQLite")
	defer storage.Close()

	assert.NotNil(t, storage.DBConn, "Expected DBConn to be initialized")
}

func TestConnectPostgres(t *testing.T) {
	// Skip this test as it requires a real PostgreSQL instance
	t.Skip("Skipping PostgreSQL connection test - requires real database instance")
}

func TestConfigureDBConnection(t *testing.T) {
	setupTestConfig(t, true)

	conn, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err, "Expected no error when connecting to SQLite for configuration test")
	err = storage.ConfigureDBConnection(conn)
	require.NoError(t, err, "Expected no error when configuring DB connection")
	sqlDB, err := conn.DB()
	require.NoError(t, err, "Expected no error when getting SQL DB from GORM connection")
	// SQLite in-memory databases have limitations on connections
	// Just verify that the connection is properly configured
	assert.NotNil(t, sqlDB, "Expected SQL DB to be available")
	assert.GreaterOrEqual(t, sqlDB.Stats().MaxOpenConnections, 1, "Expected MaxOpenConns to be at least 1")
}

func TestMigrate(t *testing.T) {
	setupTestConfig(t, true)

	// Use a fresh in-memory database for migration testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err, "Failed to open in-memory database")
	
	//nolint:reassign // DBConn is reassigned for test setup.
	storage.DBConn = db

	err = storage.Migrate()
	require.NoError(t, err, "Expected no error when migrating the database")
	assert.True(t, storage.DBConn.Migrator().HasTable(&models.Paste{}), "Expected Paste table to be created")
	
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}
}

func TestClose(t *testing.T) {
	setupTestConfig(t, true)

	err := storage.Connect(nil)
	require.NoError(t, err, "Expected no error when connecting to SQLite")
	err = storage.Close()
	require.NoError(t, err, "Expected no error when closing the database connection")
}