//go:build integration

package integration

import (
	"context"
	"testing"

	mntRepo "2_Go/internal/maintenance"

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
		t.Fatalf("failed to insert test data: %v", err)
	}

	repo := mntRepo.NewMaintenanceRepository(db)

	id, err := repo.GetLastObjId()
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

	id, err := repo.GetCurrentId()
	if err != nil {
		t.Fatalf("GetCurrentId returned error: %v", err)
	}

	if id != 456 {
		t.Errorf("expected 456, got %d", id)
	}
}

func TestIsMatchedObjectDocAndCounter(t *testing.T) {
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
	_, err = db.Exec("INSERT INTO obj_id_counter (name, obj_id) VALUES ('document', 3)")
	if err != nil {
		t.Errorf(errorInsertMsg, err)
	}

	repo := mntRepo.NewMaintenanceRepository(db)
	svc := mntRepo.NewMaintenanceService(repo)

	isMatch, err := svc.IsMatchedObjectDocAndCounter()
	if err != nil {
		t.Fatalf("\x1b[31mIsMatchedObjectDocAndCounter returned error: %v\x1b[0m", err)
	}

	if !isMatch {
		t.Errorf("\033[31mexpected true, got false\033[0m")
	}
}

func TestIsNotMatchedObjectDocAndCounter(t *testing.T) {
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
	_, err = db.Exec("INSERT INTO obj_id_counter (name, obj_id) VALUES ('document', 4)")
	if err != nil {
		t.Errorf(errorInsertMsg, err)
	}

	repo := mntRepo.NewMaintenanceRepository(db)
	svc := mntRepo.NewMaintenanceService(repo)

	isMatch, err := svc.IsMatchedObjectDocAndCounter()
	if err != nil {
		t.Fatalf("\x1b[31mIsMatchedObjectDocAndCounter returned error: %v\x1b[0m", err)
	}

	if isMatch {
		t.Errorf("\033[31mexpected false, got true\033[0m")
	}
}

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
				IssueMessage: "Test issue message",
			},
			expectedErr: false,
		},
		{
			name: "insert with not exist obj record",
			record: mntRepo.ObjVerifyRecord{
				ObjId:        6,
				HasIssue:     true,
				IssueCode:    1,
				IssueMessage: "Test issue message",
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
