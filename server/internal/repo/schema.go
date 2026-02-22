package repo

import (
	"database/sql"
	"fmt"
)

// MigrateAll runs all versioned migrations in order, skipping already-applied ones.
func MigrateAll(db *sql.DB) error {
	if err := ensureMigrationsTable(db); err != nil {
		return err
	}

	migrations := []struct {
		version int
		fn      func(*sql.DB) error
	}{
		{1, createTesterTable},
		{2, createAccountTable},
		{3, createDocumentTable},
		{4, createFilePathTable},
		{5, createObjIdTable},
		// Add new migrations here — never edit existing ones above.
		{6, func(db *sql.DB) error {
			_, err := db.Exec(`ALTER TABLE obj_doc ADD COLUMN buy_from VARCHAR(1000)`)
			return err
		}},
	}

	for _, m := range migrations {
		ran, err := hasMigrationRun(db, m.version)
		if err != nil {
			return fmt.Errorf("checking migration %d: %w", m.version, err)
		}
		if ran {
			continue
		}
		if err := m.fn(db); err != nil {
			return fmt.Errorf("migration %d failed: %w", m.version, err)
		}
		if err := markMigrationDone(db, m.version); err != nil {
			return fmt.Errorf("recording migration %d: %w", m.version, err)
		}
	}
	return nil
}

// --- helpers ---

func ensureMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version     INT PRIMARY KEY,
		applied_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	) ENGINE=InnoDB`)
	return err
}

func hasMigrationRun(db *sql.DB, version int) (bool, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, version).Scan(&count)
	return count > 0, err
}

func markMigrationDone(db *sql.DB, version int) error {
	_, err := db.Exec(`INSERT INTO schema_migrations (version) VALUES (?)`, version)
	return err
}

// --- migrations (never edit, only append new ones) ---

func createTesterTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS Tester (
		id         INT AUTO_INCREMENT PRIMARY KEY,
		name       VARCHAR(100) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	return err
}

func createAccountTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS user (
		guid        char(36),
		id          INT AUTO_INCREMENT PRIMARY KEY,
		user_name   varchar(100),
		password    varchar(255),
		created_at  timestamp default current_timestamp,
		modified_at timestamp default current_timestamp on update current_timestamp
	) ENGINE=InnoDB`)
	return err
}

func createDocumentTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS obj_doc (
		guid          char(36),
		id            INT AUTO_INCREMENT PRIMARY KEY,
		user_id       INT,
		obj_id        BIGINT,
		status        int not null,
		name_or_title varchar(100) NOT NULL,
		description   varchar(1000),
		created_at    timestamp default current_timestamp,
		modified_at   timestamp default current_timestamp on update current_timestamp,
		buy_from      varchar(1000),
		buy_price     real,
		sold_price    real,
		buy_at        timestamp,
		sold_at       timestamp,
		file_size     real,
		extension     varchar(10),
		CONSTRAINT fk_objdoc_user FOREIGN KEY (user_id) REFERENCES user(id)
	) ENGINE=InnoDB`)
	return err
}

func createFilePathTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS obj_doc_path (
		id        INT AUTO_INCREMENT PRIMARY KEY,
		doc_id    INT,
		file_path varchar(4000)
	) ENGINE=InnoDB`)
	return err
}

func createObjIdTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS obj_id_counter (
		obj_id     bigint,
		name       varchar(100) primary key,
		created_at timestamp default current_timestamp on update current_timestamp
	) ENGINE=InnoDB`)
	if err != nil {
		return fmt.Errorf("failed to create obj_id_counter: %w", err)
	}
	_, err = db.Exec(`INSERT INTO obj_id_counter (obj_id, name) VALUES (0, 'document') ON DUPLICATE KEY UPDATE obj_id = obj_id`)
	return err
}
