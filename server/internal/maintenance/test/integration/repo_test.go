//go:build integration

package integration

import (
	"database/sql"
	"os"
	"testing"

	mntRepo "2_Go/internal/maintenance"

	_ "github.com/go-sql-driver/mysql"
)

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
	_, _ = db.Exec("DROP TABLE IF EXISTS obj_doc")
	_, _ = db.Exec("CREATE TABLE obj_doc (obj_id BIGINT PRIMARY KEY)")

	// Insert test data
	_, err := db.Exec("INSERT INTO obj_doc (obj_id) VALUES (1), (2), (3)")
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
	_, _ = db.Exec("DROP TABLE IF EXISTS obj_id_counter")
	_, _ = db.Exec("CREATE TABLE obj_id_counter (name VARCHAR(255) PRIMARY KEY, obj_id BIGINT)")

	_, err := db.Exec("INSERT INTO obj_id_counter (name, obj_id) VALUES ('document', 456)")
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
	_, _ = db.Exec("DROP TABLE IF EXISTS obj_doc")
	_, _ = db.Exec("CREATE TABLE obj_doc (obj_id BIGINT PRIMARY KEY)")
	_, _ = db.Exec("DROP TABLE IF EXISTS obj_id_counter")
	_, _ = db.Exec("CREATE TABLE obj_id_counter (name VARCHAR(255) PRIMARY KEY, obj_id BIGINT)")

	// Insert test data
	_, err := db.Exec("INSERT INTO obj_doc (obj_id) VALUES (1), (2), (3)")
	if err != nil {
		t.Errorf("\033[31mfailed to insert test data: %v\033[0m", err)
	}
	_, err = db.Exec("INSERT INTO obj_id_counter (name, obj_id) VALUES ('document', 3)")
	if err != nil {
		t.Errorf("\033[31mfailed to insert test data: %v\033[0m", err)
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

	t.Logf("\033[32mSuccessfully tested IsMatchedObjectDocAndCounter\033[0m")
}
