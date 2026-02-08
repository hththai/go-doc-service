package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all application configuration.
type Config struct {
	Env      string
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	CORS     CORSConfig
	Redis    RedisConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	Name     string
}

type JWTConfig struct {
	Secret string
}

type CORSConfig struct {
	Origins []string
}

type RedisConfig struct {
	Addr string
}

var cfg *Config

// Load initializes configuration from environment files.
// It loads .env first, then overloads with environment-specific file.
func Load() *Config {
	if cfg != nil {
		return cfg
	}

	_ = godotenv.Load(".env")

	env := os.Getenv("GO_ENV")
	switch env {
	case "development":
		_ = godotenv.Overload(".env.development")
	case "production":
		_ = godotenv.Overload(".env.production")
	default:
		log.Println("GO_ENV not set, using defaults from .env")
	}

	cfg = &Config{
		Env: env,
		Server: ServerConfig{
			Port: getEnvOrDefault("SERVER_PORT", "8088"),
		},
		Database: DatabaseConfig{
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			Name:     os.Getenv("DB_NAME"),
		},
		JWT: JWTConfig{
			Secret: os.Getenv("JWT_SECRET"),
		},
		CORS: CORSConfig{
			Origins: parseCORSOrigins(os.Getenv("CORS_ORIGINS")),
		},
		Redis: RedisConfig{
			Addr: getEnvOrDefault("REDIS_ADDR", "redis-service:6379"),
		},
	}

	return cfg
}

// Get returns the current configuration.
// Panics if Load() hasn't been called.
func Get() *Config {
	if cfg == nil {
		panic("config: Load() must be called before Get()")
	}
	return cfg
}

// MustLoad is like Load but panics on missing required config.
// Useful for tests that need config initialized.
func MustLoad() *Config {
	return Load()
}

// IsProduction returns true if running in production environment.
func IsProduction() bool {
	return cfg != nil && cfg.Env == "production"
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseCORSOrigins(origins string) []string {
	if origins == "" {
		return []string{}
	}
	return strings.Split(origins, ",")
}
