package storage

import (
	"fmt"
	"time"

	"github.com/coolguy1771/wastebin/config"
	"github.com/coolguy1771/wastebin/models"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	DBConn *gorm.DB
)

// Connect to the database
func Connect() (*models.DB, error) {
	var (
		dsn  string
		conn *gorm.DB
		err  error
	)

	if config.Conf.LocalDB {
		conn, err = gorm.Open(sqlite.Open("dev.db"), &gorm.Config{})
		if err != nil {
			return nil, err
		}
	} else {
		// Create Database connection string and connect to database
		dsn = fmt.Sprintf("user=%s password=%s host=%s dbname=%s port=%d sslmode=disable", config.Conf.DBUser, config.Conf.DBPassword, config.Conf.DBHost, config.Conf.DBName, config.Conf.DBPort)
		conn, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			return nil, err
		}

		sqlDB, err := conn.DB()
		if err != nil {
			return nil, err
		}

		sqlDB.SetMaxIdleConns(config.Conf.DBMaxIdleConns)
		sqlDB.SetMaxOpenConns(config.Conf.DBMaxOpenConns)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	db := &models.DB{DB: conn}

	// Automatically migrate the database
	if err := db.AutoMigrate(&models.Paste{}); err != nil {
		db.Logger.Error("Error migrating the database", zap.Error(err))
		return nil, err
	}

	return db, nil
}

type RetryLogger struct {
	*zap.Logger
	Retries int
}

func (l *RetryLogger) Print(values ...interface{}) {
	// Check if the query failed
	if values[3] == "failed" {
		// Retry the query
		l.Logger.Info("Retrying failed query", zap.String("query", values[2].(string)), zap.Any("args", values[4:]))
		DBConn.Exec(values[2].(string), values[4:]...)
	}
}
