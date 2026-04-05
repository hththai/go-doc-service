//go:build integration

package integration

import (
	"context"
	"testing"

	mntRepo "2_Go/internal/maintenance"
	utils "2_Go/utils"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

func TestGetLastObjIdIntegration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	dropErr := dropAllTables(db)

	if dropErr != nil {
		t.Fatalf(errorDropAllTableMsg, dropErr)
	}

	_, _ = db.Exec(createObjDocTable)

	// Insert test data
	_, err := db.Exec(insertObjDocData)
	if err != nil {
		t.Fatalf(errorFailToInsertMsg, err)
	}

	repo := mntRepo.NewMaintenanceRepository(db)

	id, err := repo.GetLastObjId(context.Background())
	if err != nil {
		t.Fatalf("GetLastObjId returned error: %v", err)
	}

	if id != 3 {
		t.Errorf("expected 3, got %d", id)
	}
}

func TestGetCurrentIdIntegration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Clean table
	// _, _ = db.Exec(dropObjVerificationTable)
	// _, _ = db.Exec(dropObjIdCounterTable)

	dropErr := dropAllTables(db)

	if dropErr != nil {
		t.Fatalf(errorDropAllTableMsg, dropErr)
	}

	_, _ = db.Exec(createObjIdCounterTable)

	_, err := db.Exec(insertObjIdCounterData)
	if err != nil {
		t.Fatalf("failed to insert test data: %v", err)
	}

	repo := mntRepo.NewMaintenanceRepository(db)

	id, err := repo.GetCurrentId(context.Background())
	if err != nil {
		t.Fatalf("GetCurrentId returned error: %v", err)
	}

	if id != 456 {
		t.Errorf("expected 456, got %d", id)
	}
}

// ... existing code ...

func TestIsMatchedObjectDocAndCounter(t *testing.T) {
	tests := []struct {
		name          string
		objID         int
		wantedId      int
		expectedMatch bool
	}{
		{
			name:          "Matched",
			objID:         3,
			wantedId:      3,
			expectedMatch: true,
		},
		{
			name:          "Not Matched",
			objID:         4,
			wantedId:      3,
			expectedMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			defer db.Close()

			dropErr := dropAllTables(db)

			if dropErr != nil {
				t.Fatalf(errorDropAllTableMsg, dropErr)
			}

			_, _ = db.Exec(createObjDocTable)
			_, _ = db.Exec(createObjIdCounterTable)

			// Insert test data
			_, err := db.Exec(insertObjDocData)
			if err != nil {
				t.Errorf(errorInsertMsg, err)
			}
			_, err = db.Exec("INSERT INTO obj_id_counter (name, obj_id) VALUES ('document', ?)", tt.objID)
			if err != nil {
				t.Errorf(errorInsertMsg, err)
			}

			repo := mntRepo.NewMaintenanceRepository(db)
			svc := mntRepo.NewMaintenanceService(repo)

			isMatch, err := svc.IsMatchedObjectDocAndCounter(context.Background())
			if err != nil {
				t.Fatalf("\x1b[31mIsMatchedObjectDocAndCounter returned error: %v\x1b[0m", err)
			}

			assert.Equal(t, tt.expectedMatch, isMatch, utils.Colorize("red", "wrong return result"))
		})
	}
}

// ... existing code ...

// II Verification test
func TestInsertIssueRecord(t *testing.T) {
	tests := []struct {
		name        string
		record      mntRepo.ObjVerifyRecord
		expectedErr bool
	}{
		{
			name: "success insert",
			record: mntRepo.ObjVerifyRecord{
				ObjId:        1,
				HasIssue:     true,
				IssueCode:    1,
				IssueMessage: "Test issue message", // dummy message
			},
			expectedErr: false,
		},
		{
			name: "insert with not exist obj record",
			record: mntRepo.ObjVerifyRecord{
				ObjId:        6,
				HasIssue:     true,
				IssueCode:    -1,
				IssueMessage: "Test issue message", // dummy message
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			db := setupTestDB(t)
			defer db.Close()

			dropErr := dropAllTables(db)

			if dropErr != nil {
				t.Fatalf(errorDropAllTableMsg, dropErr)
			}
			_, _ = db.Exec(createObjDocTable)
			_, _ = db.Exec(createObjIdCounterTable)
			_, _ = db.Exec(createObjVerficiationTable)

			// Insert into obj_doc so foreign key is valid
			_, err := db.Exec(insertObjDocData)

			if err != nil {
				t.Fatalf("failed to insert obj_doc test data:")
			}

			repo := mntRepo.NewMaintenanceRepository(db)

			// Execute insert
			_, err = repo.InsertIssueRecord(context.Background(), tt.record)

			if tt.expectedErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			// Optional: verify record exists in DB
			var count int
			err = db.QueryRow("SELECT COUNT(*) FROM obj_doc_verification WHERE obj_id = ?", tt.record.ObjId).Scan(&count)
			assert.NoError(t, err)
			assert.Equal(t, 1, count)
		})
	}
}

// REPAIR TEST CASE

// Test Get all flag records.
// server/internal/maintenance/test/integration/repo_test.go
func TestGetFlagDocumentsIntegration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	dropErr := dropAllTables(db)

	if dropErr != nil {
		t.Fatalf(errorDropAllTableMsg, dropErr)
	}

	_, _ = db.Exec(createObjDocTable)
	_, _ = db.Exec(createObjVerficiationTable)

	// Insert into obj_doc so foreign key is valid
	_, _ = db.Exec(insertObjDocData)

	// Insert test data
	insertQuery := `
	INSERT INTO obj_doc_verification (obj_id, has_issue, issue_code, issue_message) VALUES 
	(1, true, 101, 'Issue message 1'),
	(2, false, 102, 'Issue message 2'),
	(3, true, 103, 'Issue message 3')
	`
	_, err := db.Exec(insertQuery)
	if err != nil {
		t.Fatalf(errorFailToInsertMsg, err)
	}

	repo := mntRepo.NewMaintenanceRepository(db)

	records, err := repo.GetFlagDocuments(context.Background())
	if err != nil {
		t.Fatalf("GetFlagDocuments returned error: %v", err)
	}

	expectedRecords := []mntRepo.ObjVerifyRecord{
		{ObjId: 1, HasIssue: true, IssueCode: 101, IssueMessage: "Issue message 1"},
		{ObjId: 3, HasIssue: true, IssueCode: 103, IssueMessage: "Issue message 3"},
	}

	if len(records) != len(expectedRecords) {
		t.Errorf("expected %d records, got %d", len(expectedRecords), len(records))
	}

	for i, record := range records {
		assert.Equal(t, expectedRecords[i].ObjId, record.ObjId)
		assert.Equal(t, expectedRecords[i].HasIssue, record.HasIssue)
		assert.Equal(t, expectedRecords[i].IssueCode, record.IssueCode)
		assert.Equal(t, expectedRecords[i].IssueMessage, record.IssueMessage)
	}
}

// Get detail flag documents.
// server/internal/maintenance/test/integration/repo_test.go
func TestGetFlagDocumentDetailIntegration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	dropErr := dropAllTables(db)

	if dropErr != nil {
		t.Fatalf(errorDropAllTableMsg, dropErr)
	}

	_, _ = db.Exec(createObjDocDetailTable)
	_, _ = db.Exec(createObjVerficiationTable)

	// Insert into obj_doc so foreign key is valid
	_, err := db.Exec(insertObjDocDetailData)

	if err != nil {
		t.Fatalf("failed to insert obj_doc test data:")
	}

	// Insert test data into obj_doc_verification
	insertQuery := `
	INSERT INTO obj_doc_verification (obj_id, has_issue) VALUES 
	(1, true),
	(2, false),
	(3, true)
	`
	_, err = db.Exec(insertQuery)
	if err != nil {
		t.Fatalf(errorFailToInsertMsg, err)
	}

	repo := mntRepo.NewMaintenanceRepository(db)

	records, err := repo.GetFlagDocumentDetail(context.Background())
	if err != nil {
		t.Fatalf("GetFlagDocumentDetail returned error: %v", err)
	}

	expectedRecords := []mntRepo.FlaggedDocument{
		{ObjId: 1, HasIssue: true},
		{ObjId: 3, HasIssue: true},
	}

	if len(records) != len(expectedRecords) {
		t.Errorf("expected %d records, got %d", len(expectedRecords), len(records))
	}

	for i, record := range records {
		assert.Equal(t, expectedRecords[i].ObjId, record.ObjId)
		assert.Equal(t, expectedRecords[i].HasIssue, record.HasIssue)
	}

}
