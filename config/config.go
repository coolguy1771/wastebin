package config

import (
	"strings"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
)

var Conf WastebinConfig

/*
	type Config struct {
		App  WebappConfig   `koanf:"APP"`
		DB   DatabaseConfig `koanf:"DB"`
		Log  LoggerConfig   `koanf:"LOG"`
		Auth AuthConfig     `koanf:"AUTH"`
	}
*/
type WastebinConfig struct {
	DBUser         string `koanf:"DB_USER"`
	DBPassword     string `koanf:"DB_PASSWORD"`
	DBHost         string `koanf:"DB_HOST"`
	DBPort         int    `koanf:"DB_PORT"`
	DBName         string `koanf:"DB_NAME"`
	DBMaxIdleConns int    `koanf:"DB_MAX_IDLE_CONNS"`
	DBMaxOpenConns int    `koanf:"DB_MAX_OPEN_CONNS"`
	WebappPort     string `koanf:"WEBAPP_PORT"`
	Dev            bool   `koanf:"DEV"`
}

type WebappConfig struct {
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
type AuthConfig struct {
}

func Load() *WastebinConfig {
	var k = koanf.New(".")
	k.Load(confmap.Provider(map[string]interface{}{
		"WEBAPP_PORT":       "3000",
		"DB_MAX_IDLE_CONNS": "10",
		"DB_MAX_OPEN_CONNS": "50",
		"DB_PORT":           "5432",
		"DB_HOST":           "localhost",
		"DB_USER":           "wastebin",
		"DB_NAME":           "wastebin",
		"LOG_LEVEL":         "INFO",
		"DEV":               "false",
	}, "."), nil)

	k.Load(env.Provider("WASTEBIN_", ".", func(s string) string {
		return strings.TrimPrefix(s, "WASTEBIN_")
	}), nil)

	k.Unmarshal("", &Conf)

	return &Conf
}
