package document

import (
	"2_Go/utils"
	"crypto/rand"
	"database/sql"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type DocumentService struct {
	repo DocumentRepository
}

func NewDocumentService(repo DocumentRepository) *DocumentService {
	return &DocumentService{repo: repo}
}

// BeginTx starts a new database transaction.
func (s *DocumentService) BeginTx() (*sql.Tx, error) {
	return s.repo.BeginTx()
}

func (s *DocumentService) SetLatestObjID(tx *sql.Tx, document *Document) (int64, error) {
	return s.repo.SetLatestObjId(tx, document)
}

func (s *DocumentService) SaveMetadataWithObjId(tx *sql.Tx, objId *int64, document *Document) (int64, error) {
	return s.repo.SaveMetadataWithObjId(tx, objId, document)
}

func (s *DocumentService) SaveDocumentMetadata(tx *sql.Tx, document *Document) (int64, error) {

	return s.repo.SaveMetadata(tx, document)
}

func (s *DocumentService) SaveFilePath(tx *sql.Tx, document *Document) error {
	return s.repo.InsertFilePath(tx, document)
}

// UploadInput contains the data needed for document upload (decoupled from HTTP layer).
// Modifying when model change to get input
type UploadInput struct {
	Title       string
	Description string
	BuyFrom     string
	BuyAt       *time.Time
	BuyPrice    string
	Items       []Item
	UserId      int
	File        *multipart.FileHeader // nil if no file attached
}

// FileSaveFunc is a function type for saving uploaded files to disk.
type FileSaveFunc func(file *multipart.FileHeader, dst string) error

var IndexFolder = 100

// UploadDocument handles the document upload logic.
// Modify when model change to get input
func (s *DocumentService) UploadDocument(input *UploadInput, saveFile FileSaveFunc) error {

	purchaseInfo := PurchaseInfo{
		BuyFrom:  input.BuyFrom,
		BuyAt:    input.BuyAt,
		BuyPrice: input.BuyPrice,
	}

	doc := Document{
		GUID:         uuid.New().String(),
		Title:        input.Title,
		Description:  input.Description,
		PurchaseInfo: purchaseInfo,
		Items:        input.Items,
		Status:       StatusFailed,
		UserId:       input.UserId,
	}

	// Begin transaction.
	tx, err := s.BeginTx()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	// Ensure rollback if something goes wrong.
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		}
	}()

	// Retrieve and update latest objID.
	objId, err := s.SetLatestObjID(tx, &doc)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("metadata save failed: %w", err)
	}

	// Handle when there is no attached file.
	if input.File == nil {
		if _, err := s.SaveMetadataWithObjId(tx, &objId, &doc); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to save metadata: %w", err)
		}
		return tx.Commit()
	}

	// Update file-related fields.
	doc.Title = input.File.Filename
	doc.FileSize = float64(input.File.Size)
	doc.Extension = filepath.Ext(input.File.Filename)
	doc.Id = strconv.FormatInt(objId, 10)

	if err := s.saveFileAndMetadata(input.File, &doc, tx, saveFile); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("cannot save file: %w", err)
	}

	// Save the doc metadata.
	_, err = s.SaveMetadataWithObjId(tx, &objId, &doc)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	return nil
}

// saveFileAndMetadata handles saving file to temp, scanning, and saving to storage.
func (s *DocumentService) saveFileAndMetadata(file *multipart.FileHeader, doc *Document, tx *sql.Tx, saveFile FileSaveFunc) error {
	tmpPath, err := saveTemp(file, file.Filename)
	if err != nil {
		return fmt.Errorf("cannot save file tmp: %w", err)
	}

	// Delete temp folder when done.
	defer os.RemoveAll(filepath.Dir(tmpPath))

	// Scan for viruses in production.
	if utils.IsProduction() {
		if err := fileScan(tmpPath); err != nil {
			return fmt.Errorf("file error with scan: %w", err)
		}
	}

	// Build upload path and save file.
	getId, err := strconv.Atoi(doc.Id)
	if err != nil {
		return fmt.Errorf("cannot convert id: %w", err)
	}

	uploadPath := buildUploadPath(getId, doc.Id, file, IndexFolder)

	// Ensure directory exists.
	if err := os.MkdirAll(filepath.Dir(uploadPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := saveFile(file, uploadPath); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	doc.FilePath = uploadPath

	// Save file path to database.
	return s.SaveFilePath(tx, doc)
}

// buildUploadPath constructs the storage path for uploaded files.
func buildUploadPath(getId int, temId string, file *multipart.FileHeader, indexFolder int) string {
	indexIdPath := getId / indexFolder
	return "./filedata/0/" + strconv.Itoa(indexIdPath) + "/" + temId + filepath.Ext(file.Filename)
}

// randomString generates a random hex string of length n*2.
func randomString(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// saveTemp saves the uploaded file to a temporary location.
func saveTemp(fileHeader *multipart.FileHeader, fileName string) (string, error) {
	safeName, err := utils.SanitizeFileName(fileName)
	if err != nil {
		return "", err
	}

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

	destPath := filepath.Join(randomDir, safeName)

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

// fileScan runs malware detection on the file.
func fileScan(tmpPath string) error {
	cmd, err := exec.Command("/usr/local/maldetect/maldet", "-a", tmpPath).CombinedOutput()
	if err != nil {
		if rmErr := os.RemoveAll(filepath.Dir(tmpPath)); rmErr != nil {
			return fmt.Errorf("maldet scan failed: %v, cleanup error: %v", err, rmErr)
		}
		return fmt.Errorf("maldet scan failed: %v, stderr: %s", err, cmd)
	}
	return nil
}
