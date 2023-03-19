package config

import (
	"log"
	"strings"

	"github.com/knadh/koanf/v2"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"go.uber.org/zap"
)

var Conf Config

/*
		type Config struct {
		WebappConfig koanf:"APP"
		DatabaseConfig koanf:"DB"
		LoggerConfig koanf:"LOG"
		AuthConfig koanf:"AUTH"
	}

}
*/
// WastebinConfig represents the configuration for the application.
type Config struct {
	DBUser         string `koanf:"DB_USER"`
	DBPassword     string `koanf:"DB_PASSWORD"`
	DBHost         string `koanf:"DB_HOST"`
	DBPort         int    `koanf:"DB_PORT"`
	DBName         string `koanf:"DB_NAME"`
	DBMaxIdleConns int    `koanf:"DB_MAX_IDLE_CONNS"`
	DBMaxOpenConns int    `koanf:"DB_MAX_OPEN_CONNS"`
	WebappPort     string `koanf:"WEBAPP_PORT"`
	Dev            bool   `koanf:"DEV"`
	LocalDB        bool   `koanf:"LOCAL_DB"`
}

type App struct {
	WebappPort int `koanf:"WEBAPP_PORT"`
}

type DatabaseConfig struct {
	DBUser         string `koanf:"DB_USER"`
	DBPassword     string `koanf:"DB_PASSWORD"`
	DBHost         string `koanf:"DB_HOST"`
	DBPort         int    `koanf:"DB_PORT"`
	DBName         string `koanf:"DB_NAME"`
	DBMaxIdleConns int    `koanf:"DB_MAX_IDLE_CONNS"`
	DBMaxOpenConns int    `koanf:"DB_MAX_OPEN_CONNS"`
}

type LoggerConfig struct {
	LogLevel string `koanf:"LOG_LEVEL"`
}

type AuthConfig struct{}

func Load() *Config {
	k := koanf.New(".")
	k.Load(confmap.Provider(map[string]interface{}{
		"WEBAPP_PORT":       "3000",
		"DB_MAX_IDLE_CONNS": "10",
		"DB_MAX_OPEN_CONNS": "50",
		"DB_PORT":           "5432",
		"DB_HOST":           "localhost",
		"DB_USER":           "wastebin",
		"DB_NAME":           "wastebin",
		"LOG_LEVEL":         "INFO",
		"LOCAL_DB":          "false",
	}, "."), nil)

	k.Load(env.Provider("WASTEBIN_", ".", func(s string) string {
		return strings.TrimPrefix(s, "WASTEBIN_")
	}), nil)

	if err := k.Unmarshal("", &Conf); err != nil {
		log.Fatal("Error loading config", zap.Error(err))
	}

	return &Conf
}
