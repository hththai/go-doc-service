package document

type DocumentService struct {
	repo DocumentRepository
}

func NewDocumentService(repo DocumentRepository) *DocumentService {
	return &DocumentService{repo: repo}
}

func (s *DocumentService) SaveDocumentMetadata(document Document) error {

	return s.repo.SaveMetadata(document)
}
