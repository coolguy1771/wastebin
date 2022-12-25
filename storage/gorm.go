package storage

import (
	"fmt"
	"time"

	"github.com/coolguy1771/wastebin/config"
	"github.com/coolguy1771/wastebin/log"
	"github.com/coolguy1771/wastebin/models"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DBConn *gorm.DB

// Connect to the database
func Connect() error {
	var (
		dsn  string
		conn *gorm.DB
		err  error
	)

	if config.Conf.LocalDB {
		log.Info("Using local database")
		conn, err = gorm.Open(sqlite.Open("dev.db"), &gorm.Config{})
		if err != nil {
			return err
		}
		DBConn = conn
		log.Info("Connected to local database")
		return nil
	}
	log.Info("Using remote database", zap.String("host", config.Conf.DBHost), zap.Int("port", config.Conf.DBPort), zap.String("name", config.Conf.DBName))
	// Create Database connection string and connect to database
	dsn = fmt.Sprintf("user=%s password=%s host=%s dbname=%s port=%d sslmode=disable", config.Conf.DBUser, config.Conf.DBPassword, config.Conf.DBHost, config.Conf.DBName, config.Conf.DBPort)
	conn, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	sqlDB, err := conn.DB()
	sqlDB.SetMaxIdleConns(config.Conf.DBMaxIdleConns)
	sqlDB.SetMaxOpenConns(config.Conf.DBMaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err != nil {
		return err
	}
	log.Info("Connected to remote database")

	log.Info("Set SQL Connection Settings", zap.Int("max_idle_conns", config.Conf.DBMaxIdleConns), zap.Int("max_open_conns", config.Conf.DBMaxOpenConns), zap.Int("conn_max_lifetime", 3600))

	DBConn = conn
	return nil
}

// Migrate the database
func Migrate() error {
	log.Info("Beginning database migration")
	err := DBConn.AutoMigrate(&models.Paste{})
	if err != nil {
		return err
	}
	log.Info("Database migration complete")
	return nil
}

// Close the database connection
func Close() error {
	sqlDB, err := DBConn.DB()
	if err != nil {
		return err
	}
	err = sqlDB.Close()
	if err != nil {
		return err
	}
	return nil
}
