package maintenance

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// I. Verify obj_doc id equals to obj_id_counter

// II. Verify total files is matched with obj_doc_path.
// 1. Find number of files in a directory file/0/
// 2. In database, obj_doc_path fine the number of records.
// 3. Verify if those number is matched.
// 4. If matched, GOOD.

// III. REPAIR if not.
// 0. Report the list of obj_doc that need to be repaired with obj_id.
// 1. Flag the file as "REPAIR" in database.
// 	[ ] 1.1 Schema: Need column as system_flag true false (false by default) for obj_doc.
// !!!2. Move the file to a temporary folder.

// If number of files less than obj_doc_path:
// 1. Flag the missing record by obj_id and report.

// If number of files more than obj_doc_path:
// 1. Check if there is duplicated number files. ex: 8.pdf 8.png 8.txt.
// 2. In obj_doc_path, find the address of obj_id 8, file_path: ./filedata/0/0/8.pdf
// 3. Verify with obj_doc: extension, file_size

// internal/maintenance/error.go

type ErrorCode int

const (
	ErrCodeNotFound ErrorCode = iota + 1
	ErrCodeInvalidInput
	ErrCodeDatabaseError
	ErrCodeFilePath
	ErrCodeVerificationFailed
	// Add more error codes as needed
)

type Error struct {
	Code    ErrorCode
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

func NewError(code ErrorCode, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

type MaintenanceService struct {
	// Provider ObjectIDProvider
	Repo MaintenanceRepository
}

func NewMaintenanceService(repo MaintenanceRepository) *MaintenanceService {
	return &MaintenanceService{
		// Provider: &DefaultObjectIDProvider{},
		Repo: repo,
	}
}

// I. Verify matched Object Counter
func (ms *MaintenanceService) IsMatchedObjectDocAndCounter() (bool, error) {
	// lastObjId, err := ms.Provider.GetLastObjId()
	lastObjId, err := ms.Repo.GetLastObjId()

	if err != nil {
		return false, err
	}

	// currentObjId, err := ms.Provider.GetCurrentObjId()
	currentObjId, err := ms.Repo.GetCurrentId()
	if err != nil {
		return false, err
	}

	return lastObjId == currentObjId, nil
}

// II. Verify total files is matched with obj_doc_path.
// 1. Find number of files in a directory file/0/
// Step 1: Count object document has obj_doc_path
func (ms *MaintenanceService) GetTotalFileRecord() (int64, error) {
	totalRecords, err := ms.Repo.CountRecordsInFilePath()
	if err != nil {
		return -1, NewError(ErrCodeDatabaseError, "error of getting total doc path")
	}

	return totalRecords, nil
}

// Step 2: Count document files in path folder file/0/
// we have a path= /filedata/0/
// If the file id is 9, it will be stored in /filedata/0/0/9.txt
// If the file id is 31 it will be stored in /filedata/0/0/31.pdf
// If the file id is 491 it will be stored in /filedata/0/0/4/491.pdf
// If the file id is 4599 it will be stored in /file/data/0/0/45/4599.png
// Build me the function that can get total files in the path provided.
func (ms *MaintenanceService) GetTotalFilesInPath(path string) (int64, error) {
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

// Step 3: check if total file path and count file path are equal.
func (ms *MaintenanceService) IsFilePathEqualToCountFile(path string) (bool, error) {
	// Sanity path
	if strings.Trim(path, " ") == "" {
		return false, NewError(ErrCodeInvalidInput, "path cannot be empty")
	}

	// Get count file record
	totalFilePath, pathErr := ms.GetTotalFilesInPath(path)
	if pathErr != nil {
		return false, pathErr
	}

	// Get total current file record
	totalFileRecord, err := ms.GetTotalFileRecord()
	if err != nil {
		return false, err
	}

	if totalFileRecord != totalFilePath {
		return false, NewError(ErrCodeVerificationFailed, "total file record and count file record are not equal")
	}

	return true, nil
}
