package document

import "database/sql"

type DocumentRepository interface {
	SaveMetadata(document Document) error
}

type documentRepositoryImpl struct {
	db *sql.DB
}

func NewDocumentRepository(db *sql.DB) DocumentRepository {
	return &documentRepositoryImpl{db: db}
}

func (r *documentRepositoryImpl) SaveMetadata(document Document) error {
	_, err := r.db.Exec(
		"INSERT INTO obj_doc (guid, name_or_title) VALUES (?,?)", document.GUID, document.Title,
	)
	return err
}
