package storage

import (
	"testing"

	"github.com/coolguy1771/wastebin/config"
	"github.com/coolguy1771/wastebin/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestConfig initializes the configuration for testing.
func setupTestConfig(localDB bool) {
	config.Conf = config.Config{
		LocalDB:        localDB,
		DBUser:         "test_user",
		DBPassword:     "test_pass",
		DBHost:         "localhost",
		DBPort:         5432,
		DBName:         "test_db",
		DBMaxIdleConns: 5,
		DBMaxOpenConns: 10,
	}
}

// TestConnectSQLite tests the SQLite connection function.
func TestConnectSQLite(t *testing.T) {
	setupTestConfig(true) // Set LocalDB to true

	err := Connect(nil)

	defer Close()

	assert.NoError(t, err, "Expected no error when connecting to SQLite")
	assert.NotNil(t, DBConn, "Expected DBConn to be initialized")
}

// TestConnectPostgres tests the PostgreSQL connection function.
func TestConnectPostgres(t *testing.T) {
	setupTestConfig(false) // Set LocalDB to false

	// Use an in-memory SQLite for mocking purposes
	config.Conf.DBHost = ":memory:"
	config.Conf.DBName = "test.db"

	// Mock the connection with SQLite to simulate PostgreSQL
	dsn := "file::memory:?cache=shared"
	conn, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	DBConn = conn

	err = Connect(nil)

	defer Close()

	assert.NoError(t, err, "Expected no error when connecting to PostgreSQL (mocked with SQLite)")
	assert.NotNil(t, DBConn, "Expected DBConn to be initialized")
}

// TestConfigureDBConnection tests the database connection pool configuration.
func TestConfigureDBConnection(t *testing.T) {
	setupTestConfig(true) // Set LocalDB to true

	conn, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err, "Expected no error when connecting to SQLite for configuration test")

	err = configureDBConnection(conn)
	assert.NoError(t, err, "Expected no error when configuring DB connection")

	sqlDB, err := conn.DB()
	assert.NoError(t, err, "Expected no error when getting SQL DB from GORM connection")
	assert.Equal(t, 5, sqlDB.Stats().Idle, "Expected MaxIdleConns to be set to 5")
	assert.Equal(t, 10, sqlDB.Stats().MaxOpenConnections, "Expected MaxOpenConns to be set to 10")
}

// TestMigrate tests the database migration function.
func TestMigrate(t *testing.T) {
	setupTestConfig(true) // Set LocalDB to true

	err := Connect(nil)
	assert.NoError(t, err, "Expected no error when connecting to SQLite")

	defer Close()

	err = Migrate()
	assert.NoError(t, err, "Expected no error when migrating the database")

	// Check if the Paste table exists
	assert.True(t, DBConn.Migrator().HasTable(&models.Paste{}), "Expected Paste table to be created")
}

// TestClose tests the database connection closure function.
func TestClose(t *testing.T) {
	setupTestConfig(true) // Set LocalDB to true

	err := Connect(nil)
	assert.NoError(t, err, "Expected no error when connecting to SQLite")

	err = Close()
	assert.NoError(t, err, "Expected no error when closing the database connection")
}
