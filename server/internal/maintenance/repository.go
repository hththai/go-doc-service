package maintenance

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// MaintenanceRepository defines the interface for maintenance operations
type MaintenanceRepository interface {
	/// GET

	GetLastObjId() (int64, error)
	GetCurrentId() (int64, error)
	GetTotalRecordsWithFilePath() (int64, error)
	GetTotalFileInStorage(string) (int64, error)

	// Repair.

	GetFlagDocuments(ctx context.Context) ([]ObjVerifyRecord, error)
	GetFlagDocumentDetail(ctx context.Context) ([]FlaggedDocument, error)

	/// INSERT

	InsertIssueRecord(context.Context, ObjVerifyRecord) (int64, error)

	// dev scan file
	ScanFileFolder() error
}

type maintenanceRepoImpl struct {
	db *sql.DB
}

func NewMaintenanceRepository(db *sql.DB) MaintenanceRepository {
	return &maintenanceRepoImpl{db: db}
}

// GetLastObjId retrieves the maximum object ID from the database
func (r *maintenanceRepoImpl) GetLastObjId() (int64, error) {
	var maxObjId int64
	err := r.db.QueryRow("SELECT MAX(obj_id) AS max_obj_id FROM obj_doc").Scan(&maxObjId)
	if err != nil {
		return 0, err
	}
	return maxObjId, nil
}

// GetCurrentId retrieves the current object ID from the database
func (r *maintenanceRepoImpl) GetCurrentId() (int64, error) {
	var currentId int64
	err := r.db.QueryRow("SELECT obj_id FROM obj_id_counter WHERE name='document'").Scan(&currentId)
	if err != nil {
		return 0, err
	}
	return currentId, nil
}

// Count the number of records with file path in database.
func (r *maintenanceRepoImpl) GetTotalRecordsWithFilePath() (int64, error) {
	var totalRecords int64
	err := r.db.QueryRow("SELECT COUNT(id) AS total_records FROM obj_doc_path").Scan(&totalRecords)
	if err != nil {
		return 0, err
	}
	return totalRecords, nil
}

// Get the number of files in file storage.
// we have a path= /filedata/0/
// If the file id is 9, it will be stored in /filedata/0/0/9.txt
// If the file id is 31 it will be stored in /filedata/0/0/31.pdf
// If the file id is 491 it will be stored in /filedata/0/0/4/491.pdf
// If the file id is 4599 it will be stored in /file/data/0/0/45/4599.png
// Build me the function that can get total files in the path provided.
func (r *maintenanceRepoImpl) GetTotalFileInStorage(path string) (int64, error) {
	var count int64

	err := filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Ignore hidden/system files
		// resolve .DS_Store file
		if strings.HasPrefix(d.Name(), ".") {
			return nil
		}

		if !d.IsDir() {
			count++
		}
		return nil
	})

	if err != nil {
		// return 0, NewError(ErrCodeFilePath, "failed walking directory")
		return 0, fmt.Errorf("failed walking directory")
	}

	return count, nil
}

// Flag the issue record
// has a new table call obj_doc_verification
// if the issue found, update to this record and reference to obj_doc table with obj_id
func (r *maintenanceRepoImpl) InsertIssueRecord(ctx context.Context, record ObjVerifyRecord) (int64, error) {
	query := `
        INSERT INTO obj_doc_verification (obj_id, has_issue, issue_code, issue_message)
        VALUES (?, ?, ?, ?)
    `
	res, err := r.db.ExecContext(ctx, query,
		record.ObjId,
		record.HasIssue,
		record.IssueCode,
		record.IssueMessage,
	)

	if err != nil {
		return 0, fmt.Errorf("InsertIssueRecord: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("InsertIssueRecord LastInsertId: %w", err)
	}

	return id, nil
}

// REPAIR STEPS

// Get flag documents
// A function to retrieve all objects in obj_doc_verification, which have the has_issue is true
// TODO: review if need to keep this function
func (r *maintenanceRepoImpl) GetFlagDocuments(ctx context.Context) ([]ObjVerifyRecord, error) {

	errMess := "GetFlagDocuments: %w"

	rows, err := r.db.QueryContext(ctx, "SELECT obj_id, has_issue, issue_code, issue_message FROM obj_doc_verification WHERE has_issue = true")
	if err != nil {
		return nil, fmt.Errorf(errMess, err)
	}
	defer rows.Close()

	var records []ObjVerifyRecord

	for rows.Next() {
		var record ObjVerifyRecord
		err := rows.Scan(&record.ObjId, &record.HasIssue, &record.IssueCode, &record.IssueMessage)
		if err != nil {
			return nil, fmt.Errorf(errMess, err)
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(errMess, err)
	}

	return records, nil
}

// Get all flagged details
func (r *maintenanceRepoImpl) GetFlagDocumentDetail(ctx context.Context) ([]FlaggedDocument, error) {

	errMess := "GetFlagDocumentDetail: %w"

	rows, err := r.db.QueryContext(ctx, `
        SELECT
            od.obj_id,
            odv.has_issue,
            od.status,
            od.name_or_title,
            od.description,
            od.created_at,
            od.modified_at,
            od.file_size,
            od.extension
        FROM obj_doc AS od
        JOIN obj_doc_verification AS odv
            ON od.obj_id = odv.obj_id
        WHERE odv.has_issue = TRUE;
    `)
	if err != nil {
		return nil, fmt.Errorf(errMess, err)
	}
	defer rows.Close()

	var records []FlaggedDocument

	for rows.Next() {
		var record FlaggedDocument
		err := rows.Scan(&record.ObjId,
			&record.HasIssue,
			&record.Status,
			&record.NameOrTitle,
			&record.Description,
			&record.CreatedAt,
			&record.ModifiedAt,
			&record.FileSize,
			&record.Extension)

		if err != nil {
			return nil, fmt.Errorf(errMess, err)
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(errMess, err)
	}

	return records, nil
}

// dummy test
func (r *maintenanceRepoImpl) ScanFileFolder() error {
	return scanFileFolder()
}

// Scan file folder from config and find the duplicate object name files.
// if there is a duplicate file name such as 8.pdf 8.png, add them into the list result
// copy the file and rename it to 8_1.pdf or 8_2.png and store in temp folder in the path ./filedata/0/temp/..
// get objId and create a temp folder ./filedata/0/temp/..
func scanFileFolder() error {
	configFilePath := os.Getenv("FILE_BASE_PATH")
	if configFilePath == "" {
		return fmt.Errorf("FILE_BASE_PATH environment variable is not set")
	}

	dirPath, err := getConfigDirectory(configFilePath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	fileMap, err := buildFileMap(dirPath)
	if err != nil {
		return fmt.Errorf("failed to build file map: %w", err)
	}

	temp := "temp"
	tempDir := filepath.Join(configFilePath, temp)

	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	for _, paths := range fileMap {
		if len(paths) > 1 {
			if err := processDuplicateFiles(paths, tempDir); err != nil {
				return fmt.Errorf("failed to process duplicate files: %w", err)
			}
		} else {
			fmt.Printf("Single file path found for key: %s\n", paths[0])
		}
	}

	return nil
}

func buildFileMap(dirPath string) (map[string][]string, error) {
	fileMap := make(map[string][]string)

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			baseName := filepath.Base(path)
			ext := filepath.Ext(baseName)
			key := baseName[:len(baseName)-len(ext)]

			fileMap[key] = append(fileMap[key], path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return fileMap, nil
}

func processDuplicateFiles(paths []string, tempDir string) error {
	for i, path := range paths {
		ext := filepath.Ext(path)
		baseName := filepath.Base(path[:len(path)-len(ext)])
		newFileName := fmt.Sprintf("%s_%d%s", baseName, i+1, ext)
		newFilePath := filepath.Join(tempDir, newFileName)

		if err := copyFile(path, newFilePath); err != nil {
			return fmt.Errorf("failed to copy file: %w", err)
		}
	}

	return nil
}

// getConfigDirectory reads the config file to get the directory path.
func getConfigDirectory(configPath string) (string, error) {
	// Implement reading config file and extracting directory path
	// This is a placeholder implementation
	return configPath, nil // Replace with actual logic
}

// copyFile copies a source file to a destination file.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	err = out.Sync()
	if err != nil {
		return err
	}

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return err
	}

	return nil
}
