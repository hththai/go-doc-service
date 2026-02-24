package document_test

import (
	"2_Go/internal/document"
	"database/sql"
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockRepo implements document.DocumentRepository for service-level tests.
type mockRepo struct {
	mock.Mock
}

func (m *mockRepo) SetLatestObjId(tx *sql.Tx, doc *document.Document) (int64, error) {
	args := m.Called(tx, doc)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockRepo) SaveMetadataWithObjId(tx *sql.Tx, objId *int64, doc *document.Document) (int64, error) {
	args := m.Called(tx, objId, doc)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockRepo) SaveMetadata(tx *sql.Tx, doc *document.Document) (int64, error) {
	args := m.Called(tx, doc)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockRepo) InsertFilePath(tx *sql.Tx, doc *document.Document) error {
	return m.Called(tx, doc).Error(0)
}

func (m *mockRepo) SaveItems(tx *sql.Tx, objId int64, docId int64, items []document.Item) error {
	return m.Called(tx, objId, docId, items).Error(0)
}

func (m *mockRepo) BeginTx() (*sql.Tx, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sql.Tx), args.Error(1)
}

// setupTx creates a sqlmock DB and begins a transaction.
func setupTx(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *sql.Tx) {
	t.Helper()
	db, dbMock, err := sqlmock.New()
	assert.NoError(t, err)
	dbMock.ExpectBegin()
	tx, err := db.Begin()
	assert.NoError(t, err)
	return db, dbMock, tx
}

// TestSaveItems_WithItems verifies items are persisted under the correct objId and docId.
func TestSaveItems_WithItems(t *testing.T) {
	items := []document.Item{
		{Name: "Apple", Quantity: "2", UnitPrice: "1.50", SubTotal: "3.00"},
		{Name: "Bread", Quantity: "1", UnitPrice: "3.00", SubTotal: "3.00"},
	}

	db, dbMock, tx := setupTx(t)
	defer db.Close()
	dbMock.ExpectCommit()

	repo := new(mockRepo)
	repo.On("BeginTx").Return(tx, nil)
	repo.On("SetLatestObjId", mock.Anything, mock.Anything).Return(int64(5), nil)
	repo.On("SaveMetadataWithObjId", mock.Anything, mock.Anything, mock.Anything).Return(int64(10), nil)
	repo.On("SaveItems", mock.Anything, int64(5), int64(10), items).Return(nil)

	svc := document.NewDocumentService(repo)
	err := svc.UploadDocument(&document.UploadInput{
		Title:  "Woolworths Receipt",
		UserId: 1,
		Items:  items,
	}, nil)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

// TestSaveItems_NilItems verifies SaveItems is still called when no items are provided.
func TestSaveItems_NilItems(t *testing.T) {
	db, dbMock, tx := setupTx(t)
	defer db.Close()
	dbMock.ExpectCommit()

	repo := new(mockRepo)
	repo.On("BeginTx").Return(tx, nil)
	repo.On("SetLatestObjId", mock.Anything, mock.Anything).Return(int64(1), nil)
	repo.On("SaveMetadataWithObjId", mock.Anything, mock.Anything, mock.Anything).Return(int64(2), nil)
	repo.On("SaveItems", mock.Anything, int64(1), int64(2), mock.Anything).Return(nil)

	svc := document.NewDocumentService(repo)
	err := svc.UploadDocument(&document.UploadInput{
		Title:  "Empty Receipt",
		UserId: 1,
	}, nil)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

// TestSaveItems_Error verifies that a SaveItems failure causes an error and rolls back.
func TestSaveItems_Error(t *testing.T) {
	db, dbMock, tx := setupTx(t)
	defer db.Close()
	dbMock.ExpectRollback()

	repo := new(mockRepo)
	repo.On("BeginTx").Return(tx, nil)
	repo.On("SetLatestObjId", mock.Anything, mock.Anything).Return(int64(1), nil)
	repo.On("SaveMetadataWithObjId", mock.Anything, mock.Anything, mock.Anything).Return(int64(2), nil)
	repo.On("SaveItems", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("db insert failed"))

	svc := document.NewDocumentService(repo)
	err := svc.UploadDocument(&document.UploadInput{
		Title:  "Receipt",
		UserId: 1,
		Items:  []document.Item{{Name: "Apple", Quantity: "1", UnitPrice: "1.00", SubTotal: "1.00"}},
	}, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save items")
	repo.AssertExpectations(t)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

// TestSaveItems_MetadataError verifies that a metadata failure prevents SaveItems from being called.
func TestSaveItems_MetadataError(t *testing.T) {
	db, dbMock, tx := setupTx(t)
	defer db.Close()
	dbMock.ExpectRollback()

	repo := new(mockRepo)
	repo.On("BeginTx").Return(tx, nil)
	repo.On("SetLatestObjId", mock.Anything, mock.Anything).Return(int64(1), nil)
	repo.On("SaveMetadataWithObjId", mock.Anything, mock.Anything, mock.Anything).Return(int64(0), errors.New("metadata insert failed"))

	svc := document.NewDocumentService(repo)
	err := svc.UploadDocument(&document.UploadInput{
		Title:  "Receipt",
		UserId: 1,
		Items:  []document.Item{{Name: "Apple", Quantity: "1", UnitPrice: "1.00", SubTotal: "1.00"}},
	}, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save metadata")
	repo.AssertNotCalled(t, "SaveItems")
	repo.AssertExpectations(t)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}
