package storage

import (
	"fmt"
	"time"

	"github.com/coolguy1771/wastebin/config"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	DBConn *gorm.DB
)

// Connect creates a connection to database
func Connect() (err error) {


	if config.Conf.Dev {
		DBConn, err = gorm.Open(sqlite.Open("dev.db"), &gorm.Config{})
		if err != nil {
			return err
		}
	
		return nil
	}

	// Create Database connection string and connect to database
	dsn := fmt.Sprintf("user=%s password=%s host=%s dbname=%s port=%d sslmode=disable", config.Conf.DBUser, config.Conf.DBPassword, config.Conf.DBHost, config.Conf.DBName, config.Conf.DBPort)
	DBConn, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	sqlDB, err := DBConn.DB()
	if err != nil {
		return err
	}

	sqlDB.SetMaxIdleConns(config.Conf.DBMaxIdleConns)
	sqlDB.SetMaxOpenConns(config.Conf.DBMaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return nil
}
