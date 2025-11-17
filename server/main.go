package main

import (
	internal "2_Go/internal"
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var documents = []internal.Document{
	{Id: "001", Title: "Example", CreatedAt: time.Now(), ModifiedAt: time.Now()},
	{Id: "002", Title: "Example2", CreatedAt: time.Now(), ModifiedAt: time.Now()},
}

func getDocs(c *gin.Context) {
	err := c.ShouldBind(&documents)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, documents)
}

func uploadHandler(c *gin.Context) {
	Title := c.PostForm("title")

	file, err := c.FormFile("file")

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	uploadPath := "./filedata/0/" + file.Filename

	if err := c.SaveUploadedFile(file, uploadPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	csvFile := "./metadata/metadata.csv"
	f, err := os.OpenFile(csvFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open CSV File"})
		return
	}

	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	// Write a row
	record := []string{uuid.New().String(), Title, time.Now().String(), file.Filename, fmt.Sprintf("%d", file.Size), uploadPath}
	if err := writer.Write(record); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write to CSV"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Filed Uploaded",
	})

}

func writeCSV(path string, header []string, content []string) error {
	fileExists := false
	if _, err := os.Stat(path); err == nil {
		fileExists = true
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return fmt.Errorf("error opening file::: %w", err)
	}

	defer f.Close()

	writer := csv.NewWriter(f)

	defer writer.Flush()

	// If file is new, writer header first
	if !fileExists {
		if err := writer.Write(header); err != nil {
			return fmt.Errorf("error writing header::: %w", err)
		}
	}

	// Append content
	if err := writer.Write(content); err != nil {
		return fmt.Errorf("Error writing content::: %w", err)
	}

	return nil
}

func main() {
	// doc := internal.CreateNewDoc("Hello")

	// fmt.Println("This is title:::", doc.Title)
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/documents", getDocs)
	r.POST("/upload", uploadHandler)

	r.Run("localhost:8088")
}
