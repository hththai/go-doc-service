package sample

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
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

	return db, nil
}

func mainTest() {
	// connString := "appuser:password@tcp(127.0.0.1:3306)/goDocument?charset=utf8mb4&parseTime=True&loc=Local"
	// db, err := sql.Open("mysql", connString)
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatalln("Error loading .env file")
	// }

	dbCredential := LoadConfig()

	db, err := InitDB(dbCredential)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	fmt.Println("Database connection successfully!")

	r := gin.Default()

	r.GET("/testers", func(c *gin.Context) {
		rows, err := db.Query("SELECT * from Tester")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var testers []Tester

		for rows.Next() {
			var tester Tester
			if err := rows.Scan(&tester.Id, &tester.Name, &tester.CreatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			testers = append(testers, tester)
		}

		c.JSON(http.StatusOK, testers)
	})

	r.POST("/testers", func(c *gin.Context) {
		var newTester Tester
		if err := c.ShouldBindJSON(&newTester); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		query := "INSERT INTO Tester (name) VALUES (?)"
		_, err := db.Exec(query, newTester.Name)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "tester added"})
	})

	fmt.Println("Serever running....")
	r.Run(":8088")
}
