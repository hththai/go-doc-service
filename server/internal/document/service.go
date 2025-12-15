package document

import "database/sql"

type DocumentService struct {
	repo DocumentRepository
}

func NewDocumentService(repo DocumentRepository) *DocumentService {
	return &DocumentService{repo: repo}
}

func (s *DocumentService) SetLatestFileID(tx *sql.Tx, document *Document) (int64, error) {
	return s.repo.SetLatestObjId(tx, document)
}

func (s *DocumentService) SaveDocumentMetadata(tx *sql.Tx, document *Document) (int64, error) {

	return s.repo.SaveMetadata(tx, document)
}

func (s *DocumentService) SaveFilePath(tx *sql.Tx, document *Document) error {
	return s.repo.InsertFilePath(tx, document)
}
