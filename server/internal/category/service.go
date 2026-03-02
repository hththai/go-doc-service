package category

import "database/sql"

type CategoryService struct {
	repo CategoryRepository
}

func NewCategoryService(repo CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) CreateCategory(userID int, name, color string) (*Category, error) {
	return s.repo.CreateCategory(userID, name, color)
}

func (s *CategoryService) GetCategoriesByUser(userID int) ([]Category, error) {
	return s.repo.GetCategoriesByUser(userID)
}

func (s *CategoryService) UpdateCategory(guid string, userID int, name, color string) error {
	return s.repo.UpdateCategory(guid, userID, name, color)
}

func (s *CategoryService) DeleteCategory(guid string, userID int) error {
	return s.repo.SoftDeleteCategory(guid, userID)
}

// SaveDocCategories and DeleteDocCategories are called from the document service
// inside the purchase save/update transaction.
func (s *CategoryService) SaveDocCategories(tx *sql.Tx, docId int64, categoryGUIDs []string, userID int) error {
	return s.repo.SaveDocCategories(tx, docId, categoryGUIDs, userID)
}

func (s *CategoryService) DeleteDocCategories(tx *sql.Tx, docId int64) error {
	return s.repo.DeleteDocCategories(tx, docId)
}

func (s *CategoryService) GetCategoriesByDocIds(docIds []int64) (map[int64][]Category, error) {
	return s.repo.GetCategoriesByDocIds(docIds)
}
