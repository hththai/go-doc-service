package document

import (
	"database/sql"
	"fmt"
	"strconv"
)

type DocumentRepository interface {
	SetLatestObjId(tx *sql.Tx, document *Document) (int64, error)
	SaveMetadata(tx *sql.Tx, document *Document) (int64, error)
	InsertFilePath(tx *sql.Tx, document *Document) error
}

type documentRepositoryImpl struct {
	db *sql.DB
}

func NewDocumentRepository(db *sql.DB) DocumentRepository {
	return &documentRepositoryImpl{db: db}
}

// Resolve objID gapless by table if there is file(s) attachment.
func (r *documentRepositoryImpl) SetLatestObjId(tx *sql.Tx, document *Document) (int64, error) {

	var objID int64

	// Lock the counter row.
	err := tx.QueryRow(`
	SELECT obj_id FROM obj_id_counter WHERE name='document' FOR UPDATE
	`).Scan(&objID)

	// fmt.Printf("the current objID::: %d\n", objID)

	if err != nil {
		return -1, fmt.Errorf("failed to lock obj_id_counter: %w", err)
	}

	_, err = tx.Exec(`
	UPDATE obj_id_counter SET obj_id = obj_id + 1 WHERE name = 'document'
	`)

	if err != nil {
		return -1, fmt.Errorf("failed to update latest obj ID: %w", err)
	}

	// Get the objID after updated.
	err = tx.QueryRow(`
	SELECT obj_id FROM obj_id_counter WHERE name='document'
		`).Scan(&objID)

	if err != nil {
		return -1, fmt.Errorf("failed to get latest obj ID: %w", err)
	}

	return objID, nil
}

// Return int objID, if not -1.
func (r *documentRepositoryImpl) SaveMetadata(tx *sql.Tx, document *Document) (int64, error) {
	// tx, err := r.db.Begin()

	// if err != nil {
	// 	return -1, err
	// }

	// result, err := r.db.Exec(
	result, err := tx.Exec(
		`INSERT INTO obj_doc (guid, name_or_title, description, file_size, extension, status)
		VALUES (?,?,?,?,?,?)`,
		document.GUID, document.Title, document.Description, document.FileSize, document.Extension, document.Status,
	)

	if err != nil {
		// tx.Rollback()
		return -1, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		// tx.Rollback()
		return -1, err
	}

	return id, nil
}

// Insert file path.
func (r *documentRepositoryImpl) InsertFilePath(tx *sql.Tx, document *Document) error {

	objId, err := strconv.ParseInt(document.Id, 10, 64)
	if err != nil {
		return err
	}

	// _, err = r.db.Exec(
	_, err = tx.Exec(
		`INSERT INTO obj_doc_path (doc_id, file_path) VALUES (?,?)`, objId, document.FilePath,
	)
	if err != nil {
		return err
	}
	return nil
}
