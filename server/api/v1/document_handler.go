package v1

import (
	"2_Go/internal/document"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func UploadDocument(c *gin.Context, service *document.DocumentService) error {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return err
	}

	// Process to scan and save file.

	// Continue to save metadata
	doc := document.Document{
		GUID:  uuid.New().String(),
		Title: file.Filename,
	}

	if err := service.SaveDocumentMetadata(doc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save metadata"})
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File Uploaded Succesful",
	})

	return nil
}
