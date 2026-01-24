package utils

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
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

// Custom Util function to create a folder and name if it is not exist.
func createFolderAndFile(folderName string, fileWithExt string) (*os.File, error) {

	err := os.MkdirAll("./"+folderName, os.ModePerm)
	if err != nil {

		return nil, errors.New("failed to create dictionary")
	}

	// Open and Create the log file.
	logFile, err := os.OpenFile("./"+folderName+"/"+fileWithExt, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {

		return nil, errors.New("failed to create file")
	}

	return logFile, nil
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

// Check validate user.
func IsValidUser(c *gin.Context) bool {
	jwtUser := c.GetString("username") // from Token
	reqUser := c.Param("username")     // from request

	if jwtUser != reqUser {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return true
	}
	return false
}

func SayHello() error {
	fmt.Println("Hello")
	return nil
}
