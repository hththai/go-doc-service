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
		{6, createItemTable},
		{7, alterDocumentPriceAndDateColumns},
		{8, alterBuyAtToDate},
		{9, addFileNameColumn},
		{10, ensureItemTable},
		{11, fixStatusFailedToActive},
		{12, widenItemNameColumn},
		{13, addUserStatusColumn},
		{14, addItemStatusColumn},
		{15, addItemGuidColumn},
		{16, backfillItemGuid},
		// Add new migrations here — never edit existing ones above.
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

func alterBuyAtToDate(db *sql.DB) error {
	_, err := db.Exec(`ALTER TABLE obj_doc MODIFY COLUMN buy_at DATE`)
	if err != nil {
		return fmt.Errorf("failed to alter buy_at to DATE: %w", err)
	}
	return nil
}

func alterDocumentPriceAndDateColumns(db *sql.DB) error {
	_, err := db.Exec(`ALTER TABLE obj_doc
		MODIFY COLUMN buy_at     DATETIME,
		MODIFY COLUMN sold_at    DATETIME,
		MODIFY COLUMN buy_price  DECIMAL(10, 2),
		MODIFY COLUMN sold_price DECIMAL(10, 2)`)
	if err != nil {
		return fmt.Errorf("failed to alter obj_doc columns: %w", err)
	}
	return nil
}

// ensureItemTable re-creates obj_item if it was skipped due to a migration version
// collision (createItemTable was renumbered from v7 to v6, so production databases
// that had the old v6 recorded in schema_migrations never ran createItemTable).
func ensureItemTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS obj_item (
		id          INT AUTO_INCREMENT PRIMARY KEY,
		obj_id      BIGINT NOT NULL,
		doc_id      INT NOT NULL,
		name        VARCHAR(255) NOT NULL,
		quantity    DECIMAL(10, 2) NOT NULL DEFAULT 0,
		price       DECIMAL(10, 2) NOT NULL DEFAULT 0,
		total       DECIMAL(10, 2) NOT NULL DEFAULT 0,
		created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		modified_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		CONSTRAINT fk_item_doc FOREIGN KEY (doc_id) REFERENCES obj_doc(id)
	) ENGINE=InnoDB`)
	if err != nil {
		return fmt.Errorf("failed to ensure obj_item: %w", err)
	}
	_, err = db.Exec(`INSERT INTO obj_id_counter (obj_id, name) VALUES (0, 'item') ON DUPLICATE KEY UPDATE obj_id = obj_id`)
	return err
}

func addFileNameColumn(db *sql.DB) error {
	_, err := db.Exec(`ALTER TABLE obj_doc ADD COLUMN file_name varchar(500) NULL AFTER name_or_title`)
	if err != nil {
		return fmt.Errorf("failed to add file_name column: %w", err)
	}
	return nil
}

func createItemTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS obj_item (
		id          INT AUTO_INCREMENT PRIMARY KEY,
		obj_id      BIGINT NOT NULL,
		doc_id      INT NOT NULL,
		name        VARCHAR(255) NOT NULL,
		quantity    DECIMAL(10, 2) NOT NULL DEFAULT 0,
		price       DECIMAL(10, 2) NOT NULL DEFAULT 0,
		total       DECIMAL(10, 2) NOT NULL DEFAULT 0,
		created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		modified_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		CONSTRAINT fk_item_doc FOREIGN KEY (doc_id) REFERENCES obj_doc(id)
	) ENGINE=InnoDB`)
	if err != nil {
		return fmt.Errorf("failed to create obj_item: %w", err)
	}
	_, err = db.Exec(`INSERT INTO obj_id_counter (obj_id, name) VALUES (0, 'item') ON DUPLICATE KEY UPDATE obj_id = obj_id`)
	return err
}

func widenItemNameColumn(db *sql.DB) error {
	_, err := db.Exec(`ALTER TABLE obj_item MODIFY COLUMN name VARCHAR(1000) NOT NULL`)
	if err != nil {
		return fmt.Errorf("failed to widen obj_item.name column: %w", err)
	}
	return nil
}

func addUserStatusColumn(db *sql.DB) error {
	_, err := db.Exec(`ALTER TABLE user ADD COLUMN status INT NOT NULL DEFAULT 1 AFTER modified_at`)
	if err != nil {
		return fmt.Errorf("failed to add status column to user: %w", err)
	}
	return nil
}

func addItemStatusColumn(db *sql.DB) error {
	_, err := db.Exec(`ALTER TABLE obj_item ADD COLUMN status INT NOT NULL DEFAULT 1 AFTER modified_at`)
	if err != nil {
		return fmt.Errorf("failed to add status column to obj_item: %w", err)
	}
	return nil
}

func backfillItemGuid(db *sql.DB) error {
	_, err := db.Exec(`UPDATE obj_item SET guid = UUID() WHERE guid IS NULL`)
	if err != nil {
		return fmt.Errorf("failed to backfill guid in obj_item: %w", err)
	}
	return nil
}

func addItemGuidColumn(db *sql.DB) error {
	_, err := db.Exec(`ALTER TABLE obj_item ADD COLUMN guid char(36) NULL AFTER id`)
	if err != nil {
		return fmt.Errorf("failed to add guid column to obj_item: %w", err)
	}
	return nil
}

// fixStatusFailedToActive corrects records that were saved with status = -1 (StatusFailed)
// due to a bug where UploadDocument defaulted to StatusFailed instead of StatusActive.
func fixStatusFailedToActive(db *sql.DB) error {
	_, err := db.Exec(`UPDATE obj_doc SET status = 1 WHERE status = -1`)
	if err != nil {
		return fmt.Errorf("failed to fix status -1 to 1: %w", err)
	}
	return nil
}
