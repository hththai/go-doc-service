package repo

import (
	"database/sql"
	"fmt"
)

// EnsureTables creates the Tester table if it doesn't exist.
func EnsureTables(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS Tester (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)
	`)
	return err
}

// CreateAccountTable creates the user account table.
func CreateAccountTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS user(
		guid char(36),
		id INT AUTO_INCREMENT PRIMARY KEY,
		user_name varchar(100),
		password varchar(255),
		created_at timestamp default current_timestamp,
		modified_at timestamp default current_timestamp on update current_timestamp)ENGINE=InnoDB
	`)
	return err
}

// CreateDocumentTable creates the obj_doc table.
func CreateDocumentTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS obj_doc (
		guid char(36),
		id INT AUTO_INCREMENT PRIMARY KEY,
		user_id INT,
		obj_id BIGINT,
		status int not null,
		name_or_title varchar(100) NOT NULL,
		description varchar(1000),
		created_at timestamp default current_timestamp,
		modified_at timestamp default current_timestamp on update current_timestamp,
		buy_from varchar(1000),
		buy_price real,
		sold_price real,
		buy_at timestamp,
		sold_at timestamp,
		file_size real,
		extension varchar(10),

		CONSTRAINT fk_objdoc_user
			FOREIGN KEY (user_id)
			REFERENCES user(id)
		)ENGINE=InnoDB;
	`)
	return err
}

// CreateFilePathTable creates the obj_doc_path table.
func CreateFilePathTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS obj_doc_path(
		id INT AUTO_INCREMENT PRIMARY KEY,
		doc_id INT,
		file_path varchar(4000)
		)ENGINE=InnoDB;
		`)
	return err
}

// CreateObjIdTable creates the obj_id_counter table.
func CreateObjIdTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS obj_id_counter (obj_id bigint, name varchar(100) primary key, created_at timestamp default current_timestamp on update current_timestamp)ENGINE=InnoDB;`)
	if err != nil {
		return fmt.Errorf("failed to create obj_id_counter: %w", err)
	}

	// Insert initial row only if it doesn't exist.
	_, err = db.Exec(`INSERT INTO obj_id_counter (obj_id, name) VALUES(0,'document') ON DUPLICATE KEY UPDATE obj_id=obj_id`)
	if err != nil {
		return fmt.Errorf("failed to initialize counter row: %w", err)
	}
	return nil
}

// MigrateAll runs all table migrations.
func MigrateAll(db *sql.DB) error {
	migrations := []func(*sql.DB) error{
		EnsureTables,
		CreateAccountTable,
		CreateDocumentTable,
		CreateFilePathTable,
		CreateObjIdTable,
	}

	for _, migrate := range migrations {
		if err := migrate(db); err != nil {
			return err
		}
	}
	return nil
}
