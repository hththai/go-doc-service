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
	FilePath  string  // obj_doc_path.file_path
	Extension string  // obj_doc.extension, with leading dot (e.g. ".pdf"); "" if NULL
	FileSize  float64 // obj_doc.file_size in bytes; 0 if NULL
}
