package document

import (
	"database/sql"
	"strconv"
)

type DocumentRepository interface {
	SaveMetadata(document Document) (int64, error)
	InsertFilePath(document *Document) error
}

type documentRepositoryImpl struct {
	db *sql.DB
}

func NewDocumentRepository(db *sql.DB) DocumentRepository {
	return &documentRepositoryImpl{db: db}
}

// Return int objID, if not -1.
func (r *documentRepositoryImpl) SaveMetadata(document Document) (int64, error) {
	tx, err := r.db.Begin()

	if err != nil {
		return -1, err
	}

	result, err := r.db.Exec(
		`INSERT INTO obj_doc (guid, name_or_title, description, file_size, extension, status)
		VALUES (?,?,?,?,?,?)`,
		document.GUID, document.Title, document.Description, document.FileSize, document.Extension, document.Status,
	)

	if err != nil {
		tx.Rollback()
		return -1, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		tx.Rollback()
		return -1, err
	}

	return id, nil
}

// Insert file path.
func (r *documentRepositoryImpl) InsertFilePath(document *Document) error {

	objId, err := strconv.ParseInt(document.Id, 10, 64)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(
		`INSERT INTO obj_doc_path (doc_id, file_path) VALUES (?,?)`, objId, document.FilePath,
	)
	if err != nil {
		return err
	}
	return nil
}
