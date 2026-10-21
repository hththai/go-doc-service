package maintenance

import "database/sql"

// Object to update verification records
type ObjVerifyRecord struct {
	ObjId        int64
	HasIssue     bool
	IssueCode    int
	IssueMessage string
}

// type struct detail for obj_doc table
type FlaggedDocument struct {
	ObjId       int64
	HasIssue    bool
	Status      string
	NameOrTitle string
	Description string
	CreatedAt   sql.NullTime
	ModifiedAt  sql.NullTime
	FileSize    int64
	Extension   string
}

// DocFileMeta is the stored file location and file metadata of one document,
// used to decide which file on disk is the real one.
type DocFileMeta struct {
	ObjId     int64
	Title     string  // obj_doc.name_or_title
	FilePath  string  // obj_doc_path.file_path
	Extension string  // obj_doc.extension, with leading dot (e.g. ".pdf"); "" if NULL
	FileSize  float64 // obj_doc.file_size in bytes; 0 if NULL
}

// IssueCode identifies a kind of problem found by VerifyUserDocuments.
type IssueCode int

const (
	// IssueDuplicateFile: other files named <obj_id>.* sit next to the stored file.
	IssueDuplicateFile IssueCode = iota + 1
	// IssueMissingFile: the stored file path does not exist on disk.
	IssueMissingFile
	// IssueMetadataMismatch: obj_doc extension or file_size differ from the stored file.
	IssueMetadataMismatch
	// IssueMultiplePaths: the document has more than one obj_doc_path row.
	IssueMultiplePaths
)

// DocIssue is one problem found on a user's document.
// It holds file names only, never server paths.
type DocIssue struct {
	ObjId      int64     `json:"objId"`
	Title      string    `json:"title"`
	Code       IssueCode `json:"code"`
	Message    string    `json:"message"`
	Repairable bool      `json:"repairable"`
	FileName   string    `json:"fileName"`
	ExtraFiles []string  `json:"extraFiles,omitempty"`
}

// IssueRef selects one issue to repair.
type IssueRef struct {
	ObjId int64     `json:"objId" binding:"required"`
	Code  IssueCode `json:"code" binding:"required"`
}

// RepairOutcome is the result of repairing one selected issue.
type RepairOutcome struct {
	ObjId   int64     `json:"objId"`
	Code    IssueCode `json:"code"`
	Message string    `json:"message"`
	// Err is the internal cause of a failure, for server logs only.
	Err error `json:"-"`
}

// RepairResult lists the selected issues that were repaired and those that were skipped.
type RepairResult struct {
	Repaired []RepairOutcome `json:"repaired"`
	Skipped  []RepairOutcome `json:"skipped"`
}
