package maintenance

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

type ObjectIDProvider interface {
	GetLastObjId() (int, error)
	GetCurrentObjId() (int, error)
}

type DefaultObjectIDProvider struct{}

func (p *DefaultObjectIDProvider) GetLastObjId() (int, error) {
	return getLastObjectIdDoc()
}
func (p *DefaultObjectIDProvider) GetCurrentObjId() (int, error) {
	return getCurrentObjectIdCounter()
}

type MaintenanceService struct {
	Provider ObjectIDProvider
	Repo     MaintenanceRepository
}

func NewMaintenanceService(repo MaintenanceRepository) *MaintenanceService {
	return &MaintenanceService{
		Provider: &DefaultObjectIDProvider{},
		Repo:     repo,
	}
}

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

// It returns the last objId of documents
func getLastObjectIdDoc() (int, error) {
	// Implement the logic to get the last object ID
	// For example:
	// lastObjId, err := getFromDatabase()
	// if err != nil {
	//     return 0, NewError(ErrCodeDatabaseError, "failed to get last object ID")
	// }
	// return lastObjId, nil
	return 0, NewError(ErrCodeNotFound, "last object ID not found")
}

// It returns the current objId records.
func getCurrentObjectIdCounter() (int, error) {
	// Implement the logic to get the current object ID
	// For example:
	// currentObjId, err := getFromCounter()
	// if err != nil {
	//     return 0, NewError(ErrCodeDatabaseError, "failed to get current object ID")
	// }
	// return currentObjId, nil
	return 0, NewError(ErrCodeNotFound, "current object ID not found")
}
