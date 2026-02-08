package repo

import (
	"2_Go/internal/config"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// InitDB initializes the database connection and runs migrations.
func InitDB(cfg *config.Config) (*sql.DB, error) {
	dbCfg := cfg.Database

	// Connect without database to create it if needed.
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/", dbCfg.User, dbCfg.Password, dbCfg.Host, dbCfg.Port)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS " + dbCfg.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}
	db.Close()

	// Reconnect using the target database.
	dsnWithDB := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbCfg.User, dbCfg.Password, dbCfg.Host, dbCfg.Port, dbCfg.Name)

	db, err = sql.Open("mysql", dsnWithDB)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database %s: %w", dbCfg.Name, err)
	}

	// Run migrations.
	if err := MigrateAll(db); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return db, nil
}
