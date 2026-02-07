package utils

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
)

func IsProduction() bool {
	return os.Getenv("GO_ENV") == "production"
}

// TODO: May need to make it more automation.
func GetConfigJWT() string {
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
	return os.Getenv("JWT_SECRET")
}

func SanitizeFileName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("filename cannot be empty")
	}

	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)

	// Replace spaces with underscores.
	base = strings.ReplaceAll(base, " ", "_")

	// Allow only letters, numbers, dash, underscore.
	reg := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	base = reg.ReplaceAllString(base, "_")

	// Collapse multiple underscores.
	base = regexp.MustCompile(`_+`).ReplaceAllString(base, "_")

	// Ensure something remains.
	if base == "" {
		base = "file"
	}

	return base + ext, nil
}

func SayHello() error {
	fmt.Println("Hello")
	return nil
}
