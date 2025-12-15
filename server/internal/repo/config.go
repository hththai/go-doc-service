package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type Tester struct {
	Id        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"create_at"`
}

func EnsureTables(db *sql.DB) error {
	_, err := db.Exec(`		
		CREATE TABLE IF NOT EXISTS Tester (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)
	`)

	return err
}

// Create OBJDOC table when init.
func CreateDocumentTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS obj_doc (
		guid char(36),
		id INT AUTO_INCREMENT PRIMARY KEY,
		status int not null,
		name_or_title varchar(100) NOT NULL,
		description varchar(1000),
		created_at timestamp default current_timestamp,
		modified_at timestamp default current_timestamp on update current_timestamp,
		buy_price real,
		sold_price real,
		buy_at timestamp,
		sold_at timestamp,
		file_size real,
		extension varchar(10)
		)
	`)

	return err
}

// Create ObjDocPath table.
func CreateFilePathTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS obj_doc_path(
		id INT AUTO_INCREMENT PRIMARY KEY,
		doc_id INT,
		file_path varchar(4000)
		)
		`)

	return err
}

// Create Record Object ID counter.
func CreateObjIdTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS obj_id_counter (obj_id bigint, name varchar(100) primary key, created_at timestamp default current_timestamp on update current_timestamp)`)

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

type DBCredential struct {
	user     string
	password string
	host     string
	port     string
	dbName   string
}

func LoadConfig() DBCredential {
	_ = godotenv.Load(".env")

	// Check GO_ENV.
	env := os.Getenv("GO_ENV")
	switch env {
	case "development":
		_ = godotenv.Overload(".env.development")
	case "production":
		_ = godotenv.Overload(".env.production")
	default:
		log.Println("GO_ENV not set, using defaults from .env")
	}

	fmt.Println(env)
	fmt.Println(os.Getenv("DB_USER"))

	return DBCredential{
		user:     os.Getenv("DB_USER"),
		password: os.Getenv("DB_PASSWORD"),
		host:     os.Getenv("DB_HOST"),
		port:     os.Getenv("DB_PORT"),
		dbName:   os.Getenv("DB_NAME"),
	}
}

func InitDB(dbCredential DBCredential) (*sql.DB, error) {
	// func InitDB(user, password, host, port, dbName string) (*sql.DB, error) {
	// Connect db.
	// dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/", user, password, host, port)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/", dbCredential.user, dbCredential.password, dbCredential.host, dbCredential.port)

	db, err := sql.Open("mysql", dsn)

	if err != nil {
		return nil, fmt.Errorf("failed to connection: %w", err)
	}

	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS " + dbCredential.dbName)

	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	db.Close()

	// Reconnect using the target database.
	dsnWithDB := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbCredential.user, dbCredential.password, dbCredential.host, dbCredential.port, dbCredential.dbName)

	db, err = sql.Open("mysql", dsnWithDB)

	if err := EnsureTables(db); err != nil {
		log.Fatalln(err)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database %s: %w", dbCredential.dbName, err)
	}

	// Create ObjDoc table.
	if err := CreateDocumentTable(db); err != nil {
		log.Fatalln(err)
		return nil, fmt.Errorf("failed to create document table: %w", err)
	}

	// Create DocPath Tabl
	if err := CreateFilePathTable(db); err != nil {
		log.Fatalln(err)
		return nil, fmt.Errorf("failed to create file path table: %w", err)
	}

	// Create ObjID Table
	if err := CreateObjIdTable(db); err != nil {
		log.Fatalln(err)
		return nil, fmt.Errorf("failed to create objId table: %w", err)
	}

	return db, nil
}
