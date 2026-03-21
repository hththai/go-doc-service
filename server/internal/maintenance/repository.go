package maintenance

import "database/sql"

// MaintenanceRepository defines the interface for maintenance operations
type MaintenanceRepository interface {
	GetLastObjId() (int64, error)
	GetCurrentId() (int64, error)
	CountRecordsInFilePath() (int64, error)
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

// Count the number of records in file path.
func (r *maintenanceRepoImpl) CountRecordsInFilePath() (int64, error) {
	var totalRecords int64
	err := r.db.QueryRow("SELECT COUNT(id) AS total_records FROM obj_doc_path").Scan(&totalRecords)
	if err != nil {
		return 0, err
	}
	return totalRecords, nil
}
