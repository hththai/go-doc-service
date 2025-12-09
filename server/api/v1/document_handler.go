package v1

import (
	"2_Go/internal/document"
	"2_Go/utils"
	"crypto/rand"
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

func UploadDocument(c *gin.Context, service *document.DocumentService) error {

	derscription := c.PostForm("description")
	formTitle := c.PostForm("name")
	// temId := c.PostForm("tempId") // This should be a result after saved in database.

	file, err := c.FormFile("file")

	// Handle when there is no attached file.
	if err != nil {
		// Continue to save metadata
		return saveMetadataOnly(formTitle, derscription, service, c)
	}

	// 1. Validate file

	// 2. Begin transaction

	// if err != nil {
	// 	return err
	// }

	doc := document.Document{
		GUID:        uuid.New().String(),
		Title:       file.Filename,
		Description: derscription,
		FileSize:    float64(file.Size),
		Extension:   filepath.Ext(file.Filename),
		Status:      -1,
	}
	// 3. Save file to storage related to 2. ID result
	objId, err := service.SaveDocumentMetadata(doc)

	if err != nil {
		return err
	}
	// 4. Update status
	doc.Id = strconv.FormatInt(objId, 10)
	doc.Status = 1
	// 5. Commit.

	// return saveFileAndMetadata(file, c, temId, derscription, service)
	return saveFileAndMetadata(file, c, &doc, service)
}

// handle save file and metadata when form submit attached file upload.
// Update file path
func saveFileAndMetadata(file *multipart.FileHeader, c *gin.Context, doc *document.Document, service *document.DocumentService) error {

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
	err = service.SaveFilePath(doc)

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
func saveMetadataOnly(formTitle string, derscription string, service *document.DocumentService, c *gin.Context) error {
	doc := document.Document{
		GUID:        uuid.New().String(),
		Title:       formTitle,
		Description: derscription,
	}

	if _, err := service.SaveDocumentMetadata(doc); err != nil {
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
