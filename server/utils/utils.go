package utils

import (
	"2_Go/internal/config"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// IsProduction returns true if running in production environment.
func IsProduction() bool {
	return config.IsProduction()
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

// Dummy example.
// Function to preview PDF
func previewPDF(c *gin.Context) {
	id := c.Query("id")
	token := c.Query("token")

	if token != "abc123" {
		c.String(http.StatusUnauthorized, "Unauthorized")
		return
	}

	filePath := filepath.Join("./filedata/0/0/", id+".pdf")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.String(http.StatusNotFound, "File not found")
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "inline; filename="+strconv.Quote(id+".pdf"))

	// Stream file.
	c.File(filePath)
}
