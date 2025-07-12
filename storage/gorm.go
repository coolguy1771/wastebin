package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/coolguy1771/wastebin/config"
	"github.com/coolguy1771/wastebin/log"
	"github.com/coolguy1771/wastebin/models"
	"github.com/coolguy1771/wastebin/observability"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DBConn *gorm.DB

// Connect initializes the database connection with retry logic.
func Connect(obs *observability.Provider) error {
	return ConnectWithRetry(3, obs)
}

// ConnectWithRetry initializes the database connection with retry logic.
func ConnectWithRetry(maxRetries int, obs *observability.Provider) error {
	var (
		conn *gorm.DB
		err  error
	)

	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Info("Attempting database connection", zap.Int("attempt", attempt), zap.Int("max_retries", maxRetries))

		if config.Conf.LocalDB {
			conn, err = connectSQLite(obs)
		} else {
			conn, err = connectPostgres(obs)
		}

		if err == nil {
			// Test the connection
			if err = testConnection(conn); err == nil {
				DBConn = conn
				log.Info("Database connection established successfully", zap.Int("attempt", attempt))
				return nil
			}
		}

		log.Warn("Database connection failed",
			zap.Int("attempt", attempt),
			zap.Int("max_retries", maxRetries),
			zap.Error(err))

		if attempt < maxRetries {
			// Exponential backoff
			backoffDuration := time.Duration(attempt*attempt) * time.Second
			log.Info("Retrying database connection", zap.Duration("backoff", backoffDuration))
			time.Sleep(backoffDuration)
		}
	}

	return fmt.Errorf("failed to connect to the database after %d attempts: %w", maxRetries, err)
}

// connectSQLite connects to a local SQLite database.
func connectSQLite(obs *observability.Provider) (*gorm.DB, error) {
	log.Info("Connecting to local SQLite database")

	// Configure GORM
	config := &gorm.Config{
		Logger: logger.Default,
	}

	conn, err := gorm.Open(sqlite.Open("dev.db"), config)
	if err != nil {
		log.Error("Error connecting to SQLite database", zap.Error(err))
		return nil, err
	}

	// Apply observability instrumentation after connection is established
	if obs != nil {
		conn = obs.InstrumentGorm(conn, config.Logger)
	}

	log.Info("Connected to local SQLite database")
	return conn, nil
}

// connectPostgres connects to a remote PostgreSQL database.
func connectPostgres(obs *observability.Provider) (*gorm.DB, error) {
	log.Info("Connecting to remote PostgreSQL database",
		zap.String("host", config.Conf.DBHost),
		zap.Int("port", config.Conf.DBPort),
		zap.String("name", config.Conf.DBName))

	// Determine SSL mode based on environment
	sslMode := "prefer"
	// if config.Conf.Dev {
	// 	sslMode = "prefer" // More lenient for development
	// }

	dsn := fmt.Sprintf("user=%s password=%s host=%s dbname=%s port=%d sslmode=%s",
		config.Conf.DBUser, config.Conf.DBPassword, config.Conf.DBHost, config.Conf.DBName, config.Conf.DBPort, sslMode)

	// Configure GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default,
	}

	conn, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		log.Error("Error connecting to PostgreSQL database", zap.Error(err))
		return nil, err
	}

	if err := configureDBConnection(conn); err != nil {
		return nil, err
	}

	// Apply observability instrumentation after connection is established
	if obs != nil {
		conn = obs.InstrumentGorm(conn, gormConfig.Logger)
	}

	log.Info("Connected to remote PostgreSQL database")
	return conn, nil
}

// configureDBConnection sets the database connection pool settings with improved defaults.
func configureDBConnection(conn *gorm.DB) error {
	sqlDB, err := conn.DB()
	if err != nil {
		log.Error("Failed to get DB from GORM connection", zap.Error(err))
		return err
	}

	// Set connection pool settings with reasonable defaults
	maxIdleConns := config.Conf.DBMaxIdleConns
	if maxIdleConns <= 0 {
		maxIdleConns = 5 // Reasonable default
	}

	maxOpenConns := config.Conf.DBMaxOpenConns
	if maxOpenConns <= 0 {
		maxOpenConns = 25 // Reasonable default
	}

	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetConnMaxLifetime(30 * time.Minute) // Reduced from 1 hour for better connection refresh
	sqlDB.SetConnMaxIdleTime(10 * time.Minute) // Close idle connections after 10 minutes

	log.Info("Database connection pool configured",
		zap.Int("max_idle_conns", maxIdleConns),
		zap.Int("max_open_conns", maxOpenConns),
		zap.Duration("conn_max_lifetime", 30*time.Minute),
		zap.Duration("conn_max_idle_time", 10*time.Minute))

	return nil
}

// Migrate performs automatic database schema migration.
func Migrate() error {
	log.Info("Starting database migration")
	if err := DBConn.AutoMigrate(&models.Paste{}); err != nil {
		log.Error("Error migrating the database", zap.Error(err))
		return err
	}
	log.Info("Database migration completed successfully")
	return nil
}

// testConnection tests the database connection
func testConnection(conn *gorm.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sqlDB, err := conn.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}

// HealthCheck performs a health check on the database connection
func HealthCheck(ctx context.Context) error {
	if DBConn == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	sqlDB, err := DBConn.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	// Check if we can perform a simple query
	var count int64
	if err := DBConn.WithContext(ctx).Model(&models.Paste{}).Count(&count).Error; err != nil {
		return fmt.Errorf("database query test failed: %w", err)
	}

	return nil
}

// Close closes the database connection gracefully.
func Close() error {
	if DBConn == nil {
		log.Info("Database connection is already nil, nothing to close")
		return nil
	}

	log.Info("Closing database connection...")

	sqlDB, err := DBConn.DB()
	if err != nil {
		log.Error("Failed to get DB from GORM connection", zap.Error(err))
		return err
	}

	// Set a timeout for closing the connection
	done := make(chan error, 1)
	go func() {
		done <- sqlDB.Close()
	}()

	select {
	case err := <-done:
		if err != nil {
			log.Error("Error closing the database connection", zap.Error(err))
			return err
		}
		log.Info("Database connection closed successfully")
		return nil
	case <-time.After(10 * time.Second):
		log.Warn("Database connection close timed out")
		return fmt.Errorf("database connection close timed out")
	}
}
