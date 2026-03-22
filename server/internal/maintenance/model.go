package maintenance

// Object to update verification records
type ObjVerifyRecord struct {
	ObjId        int64
	HasIssue     bool
	IssueCode    int
	IssueMessage string
}
