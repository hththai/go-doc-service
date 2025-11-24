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

// connString := "server=127.0.0,3306;database=sample_vault;uid=appuser;password=password"

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

func CreateDocumentTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS OBJDOC (
		guid varchar(16),
		id INT AUTO_INCREMENT PRIMARY KEY,
		name_or_title varchar(100) NOT NULL,
		description text,
		created_at timestamp default current_timestamp,
		modified_at timestamp,
		buy_price real,
		sold_price real,
		buy_at timestamp,
		sold_at timestamp,
		file_size real,
		extension varchar(20),
		file_path varchar(255)		
		)
	`)

	return err
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
		return nil, fmt.Errorf("Failed to connection: %w", err)
	}

	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS " + dbCredential.dbName)

	if err != nil {
		return nil, fmt.Errorf("Failed to create database: %w", err)
	}

	db.Close()

	// Reconnect using the target database.
	dsnWithDB := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbCredential.user, dbCredential.password, dbCredential.host, dbCredential.port, dbCredential.dbName)

	db, err = sql.Open("mysql", dsnWithDB)

	if err := EnsureTables(db); err != nil {
		log.Fatalln(err)
	}

	if err != nil {
		return nil, fmt.Errorf("Failed to connect to database %s: %w", dbCredential.dbName, err)
	}

	// Create ObjectDoc table.
	if err := CreateDocumentTable(db); err != nil {
		log.Fatalln(err)
	}

	return db, nil
}
