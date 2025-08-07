package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/coolguy1771/wastebin/config"
	"github.com/coolguy1771/wastebin/log"
	"github.com/coolguy1771/wastebin/models"
	"github.com/coolguy1771/wastebin/observability"
)

// DBConn is the global GORM database connection. It is initialized by Connect/ConnectWithRetry.
//
//nolint:gochecknoglobals // This is the global database connection, which is a singleton.
var DBConn *gorm.DB

const (
	defaultMaxIdleConns = 5
	defaultMaxOpenConns = 25
	connMaxLifetime     = 30 * time.Minute
	connMaxIdleTime     = 10 * time.Minute
	dbPingTimeout       = 5 * time.Second
)

// Connect initializes the database connection with retry logic.
func Connect(obs *observability.Provider) error {
	const defaultRetryAttempts = 3
	return ConnectWithRetry(defaultRetryAttempts, obs)
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
			testErr := testConnection(conn)
			if testErr == nil {
				DBConn = conn

				log.Info("Database connection established successfully", zap.Int("attempt", attempt))

				return nil
			}

			err = testErr
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
		Logger:                                   logger.Default,
		SkipDefaultTransaction:                   false,
		DefaultTransactionTimeout:                0,
		NamingStrategy:                           nil,
		FullSaveAssociations:                     false,
		NowFunc:                                  nil,
		DryRun:                                   false,
		PrepareStmt:                              false,
		PrepareStmtMaxSize:                       0,
		PrepareStmtTTL:                           0,
		DisableAutomaticPing:                     false,
		DisableForeignKeyConstraintWhenMigrating: false,
		IgnoreRelationshipsWhenMigrating:         false,
		DisableNestedTransaction:                 false,
		AllowGlobalUpdate:                        false,
		QueryFields:                              false,
		CreateBatchSize:                          0,
		TranslateError:                           false,
		PropagateUnscoped:                        false,
		ClauseBuilders:                           nil,
		ConnPool:                                 nil,
		Dialector:                                nil,
		Plugins:                                  nil,
	}

	conn, err := gorm.Open(sqlite.Open("dev.db"), config)
	if err != nil {
		log.Error("Error connecting to SQLite database", zap.Error(err))

		return nil, fmt.Errorf("sqlite open: %w", err)
	}

	// Apply observability instrumentation after connection is established
	if obs != nil {
		conn = obs.InstrumentGorm(conn, config.Logger)
	}

	log.Info("Connected to local SQLite database")

	return conn, nil
}

// connectPostgres establishes a connection to a remote PostgreSQL database using configuration parameters and applies connection pool settings and optional observability instrumentation.
// Returns the initialized GORM database connection or an error if the connection or configuration fails.
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
		Logger:                                   logger.Default,
		SkipDefaultTransaction:                   false,
		DefaultTransactionTimeout:                0,
		NamingStrategy:                           nil,
		FullSaveAssociations:                     false,
		NowFunc:                                  nil,
		DryRun:                                   false,
		PrepareStmt:                              false,
		PrepareStmtMaxSize:                       0,
		PrepareStmtTTL:                           0,
		DisableAutomaticPing:                     false,
		DisableForeignKeyConstraintWhenMigrating: false,
		IgnoreRelationshipsWhenMigrating:         false,
		DisableNestedTransaction:                 false,
		AllowGlobalUpdate:                        false,
		QueryFields:                              false,
		CreateBatchSize:                          0,
		TranslateError:                           false,
		PropagateUnscoped:                        false,
		ClauseBuilders:                           nil,
		ConnPool:                                 nil,
		Dialector:                                nil,
		Plugins:                                  nil,
	}

	conn, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		log.Error("Error connecting to PostgreSQL database", zap.Error(err))

		return nil, fmt.Errorf("postgres open: %w", err)
	}

	configErr := configureDBConnection(conn)
	if configErr != nil {
		return nil, fmt.Errorf("configure db connection: %w", configErr)
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

		return fmt.Errorf("gorm.DB().DB(): %w", err)
	}

	// Set connection pool settings with reasonable defaults
	maxIdleConns := config.Conf.DBMaxIdleConns
	if maxIdleConns <= 0 {
		maxIdleConns = defaultMaxIdleConns // Reasonable default
	}

	maxOpenConns := config.Conf.DBMaxOpenConns
	if maxOpenConns <= 0 {
		maxOpenConns = defaultMaxOpenConns // Reasonable default
	}

	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)
	sqlDB.SetConnMaxIdleTime(connMaxIdleTime)

	log.Info("Database connection pool configured",
		zap.Int("max_idle_conns", maxIdleConns),
		zap.Int("max_open_conns", maxOpenConns),
		zap.Duration("conn_max_lifetime", connMaxLifetime),
		zap.Duration("conn_max_idle_time", connMaxIdleTime))

	return nil
}

// ConfigureDBConnection sets the database connection pool settings with improved defaults.
func ConfigureDBConnection(conn *gorm.DB) error {
	return configureDBConnection(conn)
}

// Migrate performs automatic database schema migration.
func Migrate() error {
	log.Info("Starting database migration")

	err := DBConn.AutoMigrate(&models.Paste{})
	if err != nil {
		log.Error("Error migrating the database", zap.Error(err))

		return fmt.Errorf("auto migrate: %w", err)
	}

	log.Info("Database migration completed successfully")

	return nil
}

// testConnection tests the database connection.
func testConnection(conn *gorm.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbPingTimeout)
	defer cancel()

	sqlDB, err := conn.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	pingErr := sqlDB.PingContext(ctx)
	if pingErr != nil {
		return fmt.Errorf("failed to ping database: %w", pingErr)
	}

	return nil
}

// HealthCheck performs a health check on the database connection.
func HealthCheck(ctx context.Context) error {
	if DBConn == nil {
		return errors.New("database connection is not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, dbPingTimeout)
	defer cancel()

	sqlDB, err := DBConn.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	pingErr := sqlDB.PingContext(ctx)
	if pingErr != nil {
		return fmt.Errorf("database health check failed: %w", pingErr)
	}

	// Check if we can perform a simple query
	var count int64
	queryErr := DBConn.WithContext(ctx).Model(&models.Paste{}).Count(&count).Error
	if queryErr != nil {
		return fmt.Errorf("database query test failed: %w", queryErr)
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

		return fmt.Errorf("gorm.DB().DB(): %w", err)
	}

	// Set a timeout for closing the connection
	const connectionTimeout = 10 * time.Second
	done := make(chan error, 1)

	go func() {
		done <- sqlDB.Close()
	}()

	select {
	case closeErr := <-done:
		if closeErr != nil {
			log.Error("Error closing the database connection", zap.Error(closeErr))

			return closeErr
		}

		log.Info("Database connection closed successfully")

		return nil
	case <-time.After(connectionTimeout):
		log.Warn("Database connection close timed out")

		return errors.New("database connection close timed out")
	}
}
