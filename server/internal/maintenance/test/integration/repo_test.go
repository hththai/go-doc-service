//go:build integration

package integration

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	mntRepo "2_Go/internal/maintenance"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

const (
	errorInsertMsg       = "\033[31mfailed to insert test data: %v\033[0m"
	errorDropAllTableMsg = "\033[31mfailed to drop table: %s\033[0m"

	dropObjDocTable   = "DROP TABLE IF EXISTS obj_doc"
	createObjDocTable = "CREATE TABLE obj_doc (obj_id BIGINT PRIMARY KEY)"
	insertObjDocData  = "INSERT INTO obj_doc (obj_id) VALUES (1), (2), (3)"

	dropObjIdCounterTable   = "DROP TABLE IF EXISTS obj_id_counter"
	createObjIdCounterTable = "CREATE TABLE obj_id_counter (name VARCHAR(255) PRIMARY KEY, obj_id BIGINT)"
	insertObjIdCounterData  = "INSERT INTO obj_id_counter (name, obj_id) VALUES ('document', 456)"

	dropObjVerificationTable   = "DROP TABLE IF EXISTS obj_doc_verification"
	createObjVerficiationTable = `CREATE TABLE obj_doc_verification (
								id BIGINT AUTO_INCREMENT PRIMARY KEY,
								obj_id BIGINT NOT NULL,
								has_issue TINYINT(1) NOT NULL DEFAULT 0,
								issue_code VARCHAR(50) NULL,
								issue_message VARCHAR(255) NULL,
								verified_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
								FOREIGN KEY (obj_id) REFERENCES obj_doc(obj_id)
								)`
)

// Auto drop tables and references on startup
func dropAllTables(db *sql.DB) error {
	deps := map[string][]string{
		"obj_doc_verification": {"obj_doc"},
		"obj_doc":              {},
		"obj_id_counter":       {},
	}

	order, err := topoSort(deps)
	if err != nil {
		return err
	}

	for _, table := range order {
		if err := dropTable(db, table); err != nil {
			return err
		}
	}

	return nil
}

func topoSort(graph map[string][]string) ([]string, error) {
	inDegree := make(map[string]int)
	for node := range graph {
		inDegree[node] = 0
	}

	for _, parents := range graph {
		for _, parent := range parents {
			inDegree[parent]++
		}
	}

	queue := []string{}
	for node, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, node)
		}
	}

	var result []string

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		result = append(result, node)

		for _, parent := range graph[node] {
			inDegree[parent]--
			if inDegree[parent] == 0 {
				queue = append(queue, parent)
			}
		}
	}

	if len(result) != len(graph) {
		return nil, fmt.Errorf("circular dependency detected")
	}

	return result, nil
}

func dropTable(db *sql.DB, table string) error {
	stmt := fmt.Sprintf("DROP TABLE IF EXISTS %s;", table)
	if _, err := db.Exec(stmt); err != nil {
		return fmt.Errorf("failed to drop table %s: %w", table, err)
	}
	return nil
}

// ===================

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	connStr := os.Getenv("TEST_DB_DSN")
	if connStr == "" {
		connStr = "testuser:password@tcp(127.0.0.1:3306)/testdb?parseTime=true"
	}

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		t.Fatalf("failed to connect to DB: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping DB: %v", err)
	}

	return db
}

func TestGetLastObjIdIntegration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Clean table
	// _, _ = db.Exec(dropObjVerificationTable)
	// _, _ = db.Exec(dropObjDocTable)

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

	// Clean table
	// _, _ = db.Exec(dropObjVerificationTable)
	// _, _ = db.Exec(dropObjDocTable)
	// _, _ = db.Exec(dropObjIdCounterTable)

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

	// Clean table
	// _, _ = db.Exec(dropObjVerificationTable)
	// _, _ = db.Exec(dropObjDocTable)
	// _, _ = db.Exec(dropObjIdCounterTable)

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

			// Clean tables
			// _, _ = db.Exec(dropObjVerificationTable)
			// _, _ = db.Exec(dropObjDocTable)
			// _, _ = db.Exec(dropObjIdCounterTable)

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
