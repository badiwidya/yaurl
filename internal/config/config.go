package config

import (
	"log/slog"
	"os"
)

type Config struct {
	APP_HOST     string
	APP_PORT     string
	APP_BASE_URL string
	DB_STRING    string
	LOG_LEVEL    string
}

func New() *Config {
	return &Config{
		APP_HOST:     os.Getenv("APP_HOST"),
		APP_PORT:     os.Getenv("APP_PORT"),
		APP_BASE_URL: os.Getenv("APP_BASE_URL"),
		DB_STRING:    os.Getenv("GOOSE_DBSTRING"),
		LOG_LEVEL:    os.Getenv("LOG_LEVEL"),
	}
}

func (c *Config) GetLogLevel() slog.Level {
	switch c.LOG_LEVEL {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
