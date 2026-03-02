// Package testutil provides shared test helpers and mock implementations.
// It prevents test breakage when service dependencies change: update only
// this file instead of every test that constructs a DocumentService.
package testutil

import (
	"2_Go/internal/category"
	"2_Go/internal/document"
	"database/sql"
	"testing"
)

// NoOpCategoryRepository is a no-operation implementation of category.CategoryRepository.
// Use it in tests that don't exercise category behaviour so that service
// methods which call catRepo still compile and run without panicking.
type NoOpCategoryRepository struct{}

func (n *NoOpCategoryRepository) CreateCategory(_ int, _, _ string) (*category.Category, error) {
	return nil, nil
}
func (n *NoOpCategoryRepository) GetCategoriesByUser(_ int) ([]category.Category, error) {
	return nil, nil
}
func (n *NoOpCategoryRepository) UpdateCategory(_ string, _ int, _, _ string) error { return nil }
func (n *NoOpCategoryRepository) SoftDeleteCategory(_ string, _ int) error          { return nil }
func (n *NoOpCategoryRepository) SaveDocCategories(_ *sql.Tx, _ int64, _ []string, _ int) error {
	return nil
}
func (n *NoOpCategoryRepository) DeleteDocCategories(_ *sql.Tx, _ int64) error { return nil }
func (n *NoOpCategoryRepository) GetCategoriesByDocIds(_ []int64) (map[int64][]category.Category, error) {
	return map[int64][]category.Category{}, nil
}

// NewDocumentService creates a DocumentService with the given document repo and a
// no-op category repo. When a new dependency is added to DocumentService, update
// this helper instead of every test file.
func NewDocumentService(t testing.TB, docRepo document.DocumentRepository) *document.DocumentService {
	t.Helper()
	return document.NewDocumentService(docRepo, &NoOpCategoryRepository{})
}
