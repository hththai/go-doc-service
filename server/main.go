package main

import (
	internal "2_Go/internal"
	config "2_Go/repo"
	"crypto/rand"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
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

// randomString generates a random hex string of length n*2.
func randomString(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b) // ignore error for brevity
	return fmt.Sprintf("%x", b)
}

// createAndSaveTempFolder.
func saveTemp(fileHeader *multipart.FileHeader, fileName string) (string, error) {

	// Open the uploaded file
	src, err := fileHeader.Open()
	if err != nil {
		return "", err
	}

	defer src.Close()

	// Create a random folder under the temp dir.
	randomDir := filepath.Join(os.TempDir(), "upload-"+randomString(8))
	if err := os.MkdirAll(randomDir, 0755); err != nil {
		return "", err
	}

	// Build the full path for the file.
	destPath := filepath.Join(randomDir, fileName)

	// Create the file.
	dst, err := os.Create(destPath)
	if err != nil {
		return "", err
	}

	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	return destPath, nil

}

// Scan virus and return.
func fileScan(tmpPath string) error {
	cmd, err := exec.Command("/usr/local/maldetect/maldet", "-a", tmpPath).CombinedOutput()

	if err != nil {
		// Delete path if it fail. Delete the parent folder.
		if rmErr := os.RemoveAll(filepath.Dir(tmpPath)); rmErr != nil {
			return fmt.Errorf("maldet scan failed: %v, cleanup error: %v", err, rmErr)
		}

		return fmt.Errorf("maldet scan failed: %v, stderr : %s", err, cmd)
	}

	return nil
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

func uploadHandler(c *gin.Context) {

	log.Infof("Request upload from::: %s", c.ClientIP())

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

	// 1. Save to temp folder.

	tmpPath, err := saveTemp(file, file.Filename)
	if err != nil {

		log.Errorf("Cannot save temp %s", file.Filename)
		//log.Errorf("Cannot save temp fild %s", file.Filenam≥e)

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot save file tmp"})
		return
	}

	// 2. Scan virus and return.
	err = fileScan(tmpPath)
	if err != nil {

		log.Debugf("Virus detected::: %s", file.Filename)

		c.JSON(http.StatusInternalServerError, gin.H{"error": "File error with Scan"})
		return
	}

	// 3. Save file to file storage.
	// Save to filedata
	getId, err := strconv.Atoi(temId)

	if err != nil {

		log.Errorf("Connot convert id::: %s", temId)

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot convert id"})
		return
	}

	indexIdPath := getId / 100

	uploadPath := "./filedata/0/" + strconv.Itoa(indexIdPath) + "/" + temId + filepath.Ext(file.Filename)

	if err := c.SaveUploadedFile(file, uploadPath); err != nil {

		log.Errorf("Error save file::: %s", file.Filename)

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	log.Debugf("Save file completed::: %s", file.Filename)
	// -------

	// Delete os temp final.
	defer os.RemoveAll(filepath.Dir(tmpPath))

	// **** save metadata process
	// csvFile := "./metadata/metadata.csv"

	// fileExists := false
	// if _, err := os.Stat(csvFile); err == nil {
	// 	fileExists = true
	// }

	// f, err := os.OpenFile(csvFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open CSV file"})
	// 	return
	// }
	// defer f.Close()

	// writer := csv.NewWriter(f)
	// defer writer.Flush()

	// header := []string{"guid", "id", "nameortitle", "description", "filename", "size", "path"}

	// if !fileExists {
	// 	if err := writer.Write(header); err != nil {
	// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save header"})
	// 	}
	// }

	// // Write a row: name, description, file name, file size, upload path
	// record := []string{uuid.New().String(), temId, name, description, file.Filename, fmt.Sprintf("%d", file.Size), uploadPath}
	// if err := writer.Write(record); err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write to CSV"})
	// 	return
	// }
	// ****

	c.JSON(http.StatusOK, gin.H{
		"message":   "File uploaded successfully",
		"file_name": file.Filename,
		"file_size": file.Size,
		//"upload_path": uploadPath,
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
		return fmt.Errorf("error writing content::: %w", err)
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

var log = logrus.New()

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

	log.SetOutput(&lumberjack.Logger{
		Filename:   "./app/log/myapp.log",
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     7,
		Compress:   true,
	})

	log.SetLevel(logrus.DebugLevel)

	log.Info("******APPLICATION STARTED*******")

	// logFile, err := createFolderAndFile("App", "app.log")

	// if err != nil {
	// 	log.Fatalf("Failed to open log file: %v", err)
	// }
	// defer logFile.Close()

	// log.SetOutput(logFile)
	// // log.Println("This message goes to app.log")

	// // customLogger := log.New(os.Stdout, "MY_APP: ", log.Ldate|log.Ltime|log.Lshortfile)
	// // customLogger.Println("This message goes to standard output with a custom prefix and flags")

	// log.SetFormatter(&log.TextFormatter{
	// 	FullTimestamp: true,
	// })

	// log.SetLevel(log.DebugLevel)

	// log.Info("Application started")

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

	r.Run(":8088")
}
