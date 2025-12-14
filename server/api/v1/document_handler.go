package v1

import (
	"2_Go/internal/document"
	"2_Go/utils"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var INDEX_FOLDER = 100

func UploadDocument(c *gin.Context, service *document.DocumentService, db *sql.DB) error {

	derscription := c.PostForm("description")
	formTitle := c.PostForm("name")

	doc := document.Document{
		GUID:        uuid.New().String(),
		Title:       formTitle,
		Description: derscription,
		Status:      -1,
	}

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	// Always ensure rollback if something goes wrong.
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) //rathrow panic after rollback
		} else if err != nil {
			tx.Rollback()
		}
	}()

	file, err := c.FormFile("file")
	// Handle when there is no attached file.
	if err != nil {
		// Continue to save metadata
		// err = saveMetadataOnly(&doc, service, c)
		if err := saveMetadataOnly(&doc, service, c, tx); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("metadata save failed: %w", err)
		}

		return tx.Commit()

	}

	// 1. Validate file

	// 2. Begin transaction

	// if err != nil {
	// 	return err
	// }

	// Update related fields.
	doc.Title = file.Filename
	doc.FileSize = float64(file.Size)
	doc.Extension = filepath.Ext(file.Filename)

	// 3. Save file to storage related to 2. ID result
	objId, err := service.SaveDocumentMetadata(tx, &doc)

	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("metadata save failed: %w", err)
	}
	// 4. Update status
	doc.Id = strconv.FormatInt(objId, 10)
	doc.Status = 1
	// 5. Commit.

	// return saveFileAndMetadata(file, c, temId, derscription, service)
	if err := saveFileAndMetadata(file, c, &doc, service, tx); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("cannot save file: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	return nil
}

// handle save file and metadata when form submit attached file upload.
// Update file path
func saveFileAndMetadata(file *multipart.FileHeader, c *gin.Context, doc *document.Document, service *document.DocumentService, tx *sql.Tx) error {

	return errors.New("failed save file and metadata")

	tmpPath, err := saveTemp(file, file.Filename)

	// Delete os temp final.
	defer os.RemoveAll(filepath.Dir(tmpPath)) // Consider remove only temp file.

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot save file tmp"})
		return err
	}

	// Process to scan and save file.
	// 2. Scan virus and return.
	if utils.IsProduction() {
		err = fileScan(tmpPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "File error with Scan"})
			return err
		}
	}

	// TODO: should resolve after saved into database, and retrieve return new objID.

	// 3. Save file to file storage.
	// Save to filedata.
	getId, err := strconv.Atoi(doc.Id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot convert id"})
		return err
	}

	uploadPath := buildUploadPath(getId, doc.Id, file, INDEX_FOLDER)

	if err := c.SaveUploadedFile(file, uploadPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return err
	}

	doc.FilePath = uploadPath

	// 6. Save file path database.
	// TODO: Should start rollback and commit at this fdnction.
	err = service.SaveFilePath(tx, doc)

	if err != nil {
		return err
	}

	// Continue to save metadata
	// doc := document.Document{
	// 	GUID:        uuid.New().String(),
	// 	Title:       file.Filename,
	// 	Description: derscription,
	// 	FileSize:    float64(file.Size),
	// 	Extension:   filepath.Ext(file.Filename),
	// }

	// if _, err := service.SaveDocumentMetadata(doc); err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save metadata"})
	// 	return err
	// }

	c.JSON(http.StatusOK, gin.H{
		"message": "File Uploaded Succesful",
	})

	return nil
}

// store into file storage with index folder.
func buildUploadPath(getId int, temId string, file *multipart.FileHeader, index_folder int) string {
	indexIdPath := getId / index_folder

	uploadPath := "./filedata/0/" + strconv.Itoa(indexIdPath) + "/" + temId + filepath.Ext(file.Filename)
	return uploadPath
}

// Save MetadaOnly if there is no attached file upload.
func saveMetadataOnly(doc *document.Document, service *document.DocumentService, c *gin.Context, tx *sql.Tx) error {

	if _, err := service.SaveDocumentMetadata(tx, doc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save metadata"})
		return err
	}

	c.JSON(http.StatusOK, gin.H{"message": "Success"})

	return nil
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

		return nil, errors.New("Failed to create dictionary")
	}

	// Open and Create the log file.
	logFile, err := os.OpenFile("./"+folderName+"/"+fileWithExt, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {

		return nil, errors.New("Failed to create file.")
	}

	return logFile, nil
}
