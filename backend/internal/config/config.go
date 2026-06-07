package config

import (
	"fmt"
	"os"
)

type Config struct {
	AppEnv    string
	LogLevel  string
	Port      string
	JWTSecret string
	DBHost    string
	DBPort    string
	DBUser    string
	DBPass    string
	DBName    string
}

func Load() Config {
	cfg := Config{
		AppEnv:    getOrDefault("APP_ENV", "development"),
		LogLevel:  getOrDefault("LOG_LEVEL", "info"),
		Port:      getOrDefault("PORT", "8130"),
		JWTSecret: getOrDefault("JWT_SECRET", "change-me"),
		DBHost:    getOrDefault("DB_HOST", "127.0.0.1"),
		DBPort:    getOrDefault("DB_PORT", "3306"),
		DBUser:    getOrDefault("DB_USER", "root"),
		DBPass:    getOrDefault("DB_PASSWORD", "root"),
		DBName:    getOrDefault("DB_NAME", "quizlab"),
	}
	return cfg
}

func (c Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.DBUser,
		c.DBPass,
		c.DBHost,
		c.DBPort,
		c.DBName,
	)
}

func getOrDefault(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}
