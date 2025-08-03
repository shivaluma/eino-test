package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Database DatabaseConfig
	JWT      JWTConfig
	Server   ServerConfig
}

type DatabaseConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	Database     string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  time.Duration
}

type JWTConfig struct {
	AccessSecret      string
	RefreshSecret     string
	AccessExpiration  time.Duration
	RefreshExpiration time.Duration
}

type ServerConfig struct {
	Port string
	Host string
}

func Load() *Config {
	return &Config{
		Database: DatabaseConfig{
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         getEnvAsInt("DB_PORT", 5432),
			User:         getEnv("DB_USER", "postgres"),
			Password:     getEnv("DB_PASSWORD", "postgres"),
			Database:     getEnv("DB_NAME", "food_agent"),
			SSLMode:      getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns: getEnvAsInt("DB_MAX_OPEN_CONNS", 10),
			MaxIdleConns: getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			MaxLifetime:  getEnvAsDuration("DB_MAX_LIFETIME", time.Hour),
		},
		JWT: JWTConfig{
			AccessSecret:      getEnv("JWT_ACCESS_SECRET", "your-secret-key"),
			RefreshSecret:     getEnv("JWT_REFRESH_SECRET", "your-refresh-secret-key"),
			AccessExpiration:  getEnvAsDuration("JWT_ACCESS_EXPIRATION", 15*time.Minute),
			RefreshExpiration: getEnvAsDuration("JWT_REFRESH_EXPIRATION", 7*24*time.Hour),
		},
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "localhost"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(name string, defaultVal int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}

func getEnvAsDuration(name string, defaultVal time.Duration) time.Duration {
	valueStr := getEnv(name, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultVal
}