package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all application configuration.
type Config struct {
	Env              string
	Server           ServerConfig
	Database         DatabaseConfig
	JWT              JWTConfig
	CORS             CORSConfig
	Redis            RedisConfig
	RateLimit        RateLimitConfig
	Cookie           CookieConfig
	Anthropic        AnthropicConfig
	FileDataBasePath string
}

type AnthropicConfig struct {
	APIKey string
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

type RateLimitConfig struct {
	Requests int
}

type CookieConfig struct {
	Secure bool
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
	if env != "" {
		// Load .env.{GO_ENV} — works for any environment:
		// development, staging, production, etc.
		if err := godotenv.Overload(".env." + env); err != nil {
			log.Printf("No .env.%s file found, using defaults from .env", env)
		}
	} else {
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
		RateLimit: RateLimitConfig{
			Requests: getRateLimitRequests(),
		},
		Cookie: CookieConfig{
			Secure: getCookieSecure(env),
		},
		Anthropic: AnthropicConfig{
			APIKey: os.Getenv("ANTHROPIC_API_KEY"),
		},
		FileDataBasePath: getEnvOrDefault("FILE_DATA_BASE_PATH", "./filedata/0/"),
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

func getCookieSecure(env string) bool {
	if v := os.Getenv("COOKIE_SECURE"); v != "" {
		return v == "true"
	}
	// Secure by default, only disable for development
	return env != "development"
}

func parseCORSOrigins(origins string) []string {
	if origins == "" {
		return []string{}
	}
	return strings.Split(origins, ",")
}

func getRateLimitRequests() int {
	if n, err := strconv.Atoi(os.Getenv("RATE_LIMIT_REQUESTS")); err == nil && n > 0 {
		return n
	}
	return 10
}
