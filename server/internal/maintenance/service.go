package maintenance

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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
// 	[x] 1.1 Schema: Need column as system_flag true false (false by default) for obj_doc.
// !!!2. Move the file to a temporary folder.

// If number of files less than obj_doc_path:
// 1. Flag the missing record by obj_id and report.

// If number of files more than obj_doc_path:
// 1. Check if there is duplicated number files. ex: 8.pdf 8.png 8.txt.
// 2. In obj_doc_path, find the address of obj_id 8, file_path: ./filedata/0/0/8.pdf
// 3. Verify with obj_doc: extension, file_size

// IV. FILE ISSUE:
// 1. If there is duplicated files:
// 1.1 Copy them to temp file, and renamed it to refix -1 -2, 8-1.png, 8-2.pdf, etc

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

	basePath      string // root of document storage, e.g. ./filedata/0/
	quarantineDir string // where repair moves extra files
}

func NewMaintenanceService(repo MaintenanceRepository, basePath, quarantineDir string) *MaintenanceService {
	return &MaintenanceService{
		// Provider: &DefaultObjectIDProvider{},
		Repo:          repo,
		basePath:      basePath,
		quarantineDir: quarantineDir,
	}
}

// I. Verify matched Object Counter
func (ms *MaintenanceService) IsMatchedObjectDocAndCounter(ctx context.Context) (bool, error) {
	lastObjId, err := ms.Repo.GetLastObjId(ctx)
	if err != nil {
		return false, err
	}

	currentObjId, err := ms.Repo.GetCurrentId(ctx)
	if err != nil {
		return false, err
	}

	return lastObjId == currentObjId, nil
}

// II. Verify total files is matched with obj_doc_path.
// 1. Find number of files in a directory file/0/
// Step 1: Count object document has obj_doc_path
func (ms *MaintenanceService) GetTotalFileRecord(ctx context.Context) (int64, error) {
	totalRecords, err := ms.Repo.GetTotalRecordsWithFilePath(ctx)
	if err != nil {
		return -1, NewError(ErrCodeDatabaseError, "error of getting total doc path records")
	}

	return totalRecords, nil
}

// Step 2: Count document files in path folder file/0/

func (ms *MaintenanceService) GetTotalFileInStorage(path string) (int64, error) {
	count, err := ms.Repo.GetTotalFileInStorage(path)
	if err != nil {
		return 0, err
	}

	return count, nil

}

// Step 3: check if total file path and count file path are equal.
func (ms *MaintenanceService) IsFilePathEqualToCountFile(ctx context.Context, path string) (bool, error) {
	// Sanity path
	if strings.Trim(path, " ") == "" {
		return false, NewError(ErrCodeInvalidInput, "path cannot be empty")
	}

	// Get count file record
	totalFilePath, pathErr := ms.GetTotalFileInStorage(path)
	if pathErr != nil {
		return false, pathErr
	}

	// Get total current file record
	totalFileRecord, err := ms.GetTotalFileRecord(ctx)
	if err != nil {
		return false, err
	}

	if totalFileRecord != totalFilePath {
		return false, NewError(ErrCodeVerificationFailed, "total file record and count file record are not equal")
	}

	return true, nil
}

// Step 4: report the list of flagged file
// isFlag true means flag the issue doc object
func FlagIssueDocObj() error {
	return nil
}

// REPAIR PROCESS MANUALLY
// Path: maintenance/service.go
// Show flag report and manually fix on client web to decide remove flagged docs.
// there is using transition tables.

// Step 1: show flag document.
func GetFlagDocuments() error {

	return nil
}

// Step 2: list all flag document obj_id, name or title, file size, created at,extensions, and file path.

// USER VERIFY AND REPAIR
// A user verifies their own documents, sees the issues, and picks which to repair.

// foundIssue is a DocIssue plus the server-side details needed to repair it.
type foundIssue struct {
	DocIssue
	storedPath string
	extraPaths []string
	diskExt    string
	diskSize   int64
}

// maxIssueMessageLen matches obj_doc_verification.issue_message VARCHAR(255).
const maxIssueMessageLen = 255

// VerifyUserDocuments checks the user's active documents against the files on disk.
func (ms *MaintenanceService) VerifyUserDocuments(ctx context.Context, userID int) ([]DocIssue, error) {
	found, err := ms.findIssues(ctx, userID)
	if err != nil {
		return nil, err
	}

	issues := make([]DocIssue, 0, len(found))
	for _, f := range found {
		issues = append(issues, f.DocIssue)
	}
	return issues, nil
}

// RepairUserDocuments repairs the selected issues of the user's documents.
// It verifies again first, so only issues that still exist, are repairable and belong
// to the user are repaired; every other selection is returned as skipped.
func (ms *MaintenanceService) RepairUserDocuments(ctx context.Context, userID int, refs []IssueRef) (RepairResult, error) {
	found, err := ms.findIssues(ctx, userID)
	if err != nil {
		return RepairResult{}, err
	}

	current := make(map[IssueRef]foundIssue, len(found))
	for _, f := range found {
		current[IssueRef{ObjId: f.ObjId, Code: f.Code}] = f
	}

	result := RepairResult{Repaired: []RepairOutcome{}, Skipped: []RepairOutcome{}}
	done := make(map[IssueRef]bool, len(refs))

	for _, ref := range refs {
		if done[ref] {
			continue
		}
		done[ref] = true

		f, ok := current[ref]
		if !ok {
			result.Skipped = append(result.Skipped, RepairOutcome{ObjId: ref.ObjId, Code: ref.Code,
				Message: "issue no longer found; verify again"})
			continue
		}
		if !f.Repairable {
			result.Skipped = append(result.Skipped, RepairOutcome{ObjId: ref.ObjId, Code: ref.Code,
				Message: f.Message})
			continue
		}

		message, err := ms.repairIssue(ctx, userID, f)
		if err != nil {
			result.Skipped = append(result.Skipped, RepairOutcome{ObjId: ref.ObjId, Code: ref.Code,
				Message: "repair failed", Err: err})
			continue
		}

		outcome := RepairOutcome{ObjId: ref.ObjId, Code: ref.Code, Message: message}
		// The repair is done; a failed audit record is reported in Err but does not undo it.
		if _, err := ms.Repo.InsertIssueRecord(ctx, ObjVerifyRecord{
			ObjId:        f.ObjId,
			HasIssue:     false,
			IssueCode:    int(f.Code),
			IssueMessage: truncate(message, maxIssueMessageLen),
		}); err != nil {
			outcome.Err = fmt.Errorf("record repair: %w", err)
		}
		result.Repaired = append(result.Repaired, outcome)
	}

	return result, nil
}

// findIssues returns every issue found on the user's active documents that have a file.
func (ms *MaintenanceService) findIssues(ctx context.Context, userID int) ([]foundIssue, error) {
	metas, err := ms.Repo.GetDocFileMeta(ctx, userID)
	if err != nil {
		return nil, NewError(ErrCodeDatabaseError, "error of getting document file metadata")
	}

	// Group path rows by document, keeping the obj_id order from the repository.
	var order []int64
	byDoc := make(map[int64][]DocFileMeta)
	for _, m := range metas {
		if _, ok := byDoc[m.ObjId]; !ok {
			order = append(order, m.ObjId)
		}
		byDoc[m.ObjId] = append(byDoc[m.ObjId], m)
	}

	var issues []foundIssue
	for _, objId := range order {
		docIssues, err := ms.checkDocument(byDoc[objId])
		if err != nil {
			return nil, err
		}
		issues = append(issues, docIssues...)
	}
	return issues, nil
}

// checkDocument returns the issues of one document, given all its path rows.
func (ms *MaintenanceService) checkDocument(rows []DocFileMeta) ([]foundIssue, error) {
	m := rows[0]
	base := DocIssue{ObjId: m.ObjId, Title: m.Title, FileName: filepath.Base(m.FilePath)}

	// With several paths it is unknown which file is real, so nothing else is checked.
	if len(rows) > 1 {
		issue := base
		issue.Code = IssueMultiplePaths
		issue.Message = fmt.Sprintf("document has %d stored file paths; contact support", len(rows))
		return []foundIssue{{DocIssue: issue}}, nil
	}

	info, err := os.Stat(m.FilePath)
	if os.IsNotExist(err) {
		issue := base
		issue.Code = IssueMissingFile
		issue.Message = "file is missing; re-upload it from Edit"
		return []foundIssue{{DocIssue: issue}}, nil
	}
	if err != nil {
		return nil, NewError(ErrCodeFilePath, fmt.Sprintf("cannot read file of document %d", m.ObjId))
	}

	var issues []foundIssue

	siblings, err := findSiblingFiles(filepath.Dir(m.FilePath), strconv.FormatInt(m.ObjId, 10))
	if err != nil {
		return nil, NewError(ErrCodeFilePath, fmt.Sprintf("cannot read folder of document %d", m.ObjId))
	}
	var extraPaths, extraNames []string
	for _, p := range siblings {
		if filepath.Base(p) == base.FileName {
			continue
		}
		extraPaths = append(extraPaths, p)
		extraNames = append(extraNames, filepath.Base(p))
	}
	if len(extraPaths) > 0 {
		issue := base
		issue.Code = IssueDuplicateFile
		issue.Repairable = true
		issue.ExtraFiles = extraNames
		issue.Message = fmt.Sprintf("%d extra file(s) next to %s", len(extraNames), base.FileName)
		issues = append(issues, foundIssue{DocIssue: issue, storedPath: m.FilePath, extraPaths: extraPaths})
	}

	diskExt := filepath.Ext(m.FilePath)
	if normalizeExt(m.Extension) != normalizeExt(diskExt) || int64(m.FileSize) != info.Size() {
		issue := base
		issue.Code = IssueMetadataMismatch
		issue.Repairable = true
		issue.Message = fmt.Sprintf("saved as %q, %d bytes; file is %q, %d bytes",
			m.Extension, int64(m.FileSize), diskExt, info.Size())
		issues = append(issues, foundIssue{DocIssue: issue, storedPath: m.FilePath, diskExt: diskExt, diskSize: info.Size()})
	}

	return issues, nil
}

// repairIssue applies the repair for one issue and returns a message describing it.
func (ms *MaintenanceService) repairIssue(ctx context.Context, userID int, f foundIssue) (string, error) {
	switch f.Code {
	case IssueDuplicateFile:
		var moved []string
		for _, p := range f.extraPaths {
			if !isWithin(ms.basePath, p) || filepath.Clean(p) == filepath.Clean(f.storedPath) {
				return "", fmt.Errorf("refusing to move %s", p)
			}
			if _, err := moveToQuarantine(p, ms.quarantineDir, f.ObjId); err != nil {
				return "", err
			}
			moved = append(moved, filepath.Base(p))
		}
		return fmt.Sprintf("moved %s to quarantine", strings.Join(moved, ", ")), nil

	case IssueMetadataMismatch:
		if err := ms.Repo.UpdateFileMetadata(ctx, f.ObjId, userID, f.diskExt, float64(f.diskSize)); err != nil {
			return "", err
		}
		return fmt.Sprintf("updated metadata to %q, %d bytes", f.diskExt, f.diskSize), nil
	}

	return "", fmt.Errorf("issue code %d is not repairable", f.Code)
}

// normalizeExt makes ".PDF", "pdf" and ".pdf" compare equal.
func normalizeExt(ext string) string {
	return strings.ToLower(strings.TrimPrefix(ext, "."))
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}
