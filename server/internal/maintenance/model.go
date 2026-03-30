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
