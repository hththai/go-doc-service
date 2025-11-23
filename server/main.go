package main

import (
	internal "2_Go/internal"
	config "2_Go/repo"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
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
	name := c.PostForm("name")
	description := c.PostForm("description")
	temId := c.PostForm("tempId")

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	// rename the file to id file.
	// Resolve the stored file path.
	// uploadPath := "./files/" + file.Filename

	getId, err := strconv.Atoi(temId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot convert id"})
		return
	}

	indexIdPath := getId / 100

	uploadPath := "./filedata/0/" + strconv.Itoa(indexIdPath) + "/" + temId + filepath.Ext(file.Filename)
	if err := c.SaveUploadedFile(file, uploadPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	csvFile := "./metadata/metadata.csv"

	fileExists := false
	if _, err := os.Stat(csvFile); err == nil {
		fileExists = true
	}

	f, err := os.OpenFile(csvFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open CSV file"})
		return
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	header := []string{"guid", "id", "nameortitle", "description", "filename", "size", "path"}

	if !fileExists {
		if err := writer.Write(header); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save header"})
		}
	}

	// Write a row: name, description, file name, file size, upload path
	record := []string{uuid.New().String(), temId, name, description, file.Filename, fmt.Sprintf("%d", file.Size), uploadPath}
	if err := writer.Write(record); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write to CSV"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "File uploaded successfully",
		"file_name":   file.Filename,
		"file_size":   file.Size,
		"upload_path": uploadPath,
		"name":        name,
		"description": description,
		"id":          temId,
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

// Custom Util function to create a folder and name if it is not exist.
func createFolderAndFile(folderName string, fileWithExt string) (*os.File, error) {

	err := os.MkdirAll("./"+folderName, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create dictionary: %v", err)
		return nil, errors.New("Failed to create dictionary")
	}

	// Open and Create the log file.
	logFile, err := os.OpenFile("./"+folderName+"/"+fileWithExt, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
		return nil, errors.New("Failed to create file.")
	}

	return logFile, nil
}

func main() {

	dbCredential := config.LoadConfig()

	db, err := config.InitDB(dbCredential)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	fmt.Println("Database connection successfully!")
	// doc := internal.CreateNewDoc("Hello")
	// **************EXAMPLE LOG**************
	logFile, err := createFolderAndFile("App", "app.log")

	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)
	log.Println("This message goes to app.log")

	customLogger := log.New(os.Stdout, "MY_APP: ", log.Ldate|log.Ltime|log.Lshortfile)
	customLogger.Println("This message goes to standard output with a custom prefix and flags")

	// *******************************
	// fmt.Println("This is title:::", doc.Title)
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/documents", getDocs)
	r.POST("/upload", uploadHandler)

	r.GET("/preview", previewPDF)

	r.Run("localhost:8088")
}
