package config

import (
	internal "2_Go/internal/document"
	"database/sql"

	"github.com/google/uuid"
)

type DocObjectRepo interface {
	SaveDoc(db *sql.DB, document internal.Document) error
	GetDoc() error
}

type MySQLDocRepo struct{}

func (r *MySQLDocRepo) SaveDoc(db *sql.DB, document internal.Document) error {
	_, err := db.Exec("INSERT INTO obj_doc (guid, name_or_title) VALUES(?, ?)", uuid.New().String(), document.Title)

	return err
}

func (r *MySQLDocRepo) GetDoc() error {
	return nil
}
