package maintenance

import (
	"database/sql"
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
		return 0, NewError(ErrCodeFilePath, "failed walking directory")
	}

	return count, nil
}
