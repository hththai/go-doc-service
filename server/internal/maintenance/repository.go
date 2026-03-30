package maintenance

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

// MaintenanceRepository defines the interface for maintenance operations
type MaintenanceRepository interface {
	GetLastObjId() (int64, error)
	GetCurrentId() (int64, error)
	GetTotalRecordsWithFilePath() (int64, error)
	GetTotalFileInStorage(string) (int64, error)

	// Repair.
	GetFlagDocuments(ctx context.Context) ([]ObjVerifyRecord, error)

	InsertIssueRecord(context.Context, ObjVerifyRecord) (int64, error)
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
