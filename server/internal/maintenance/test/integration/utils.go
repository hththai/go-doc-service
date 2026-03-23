package integration

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
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
