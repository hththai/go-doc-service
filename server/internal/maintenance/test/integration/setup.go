package integration

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
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
