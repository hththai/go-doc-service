package document_test

import (
	"2_Go/internal/document"
	"2_Go/internal/testutil"
	"database/sql"
	"errors"
	"mime/multipart"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockInt64Result extracts (int64, error) from a testify mock call result.
func mockInt64Result(args mock.Arguments) (int64, error) {
	return args.Get(0).(int64), args.Error(1)
}

// mockRepo implements document.DocumentRepository for service-level tests.
type mockRepo struct {
	mock.Mock
}

func (m *mockRepo) SetLatestObjId(tx *sql.Tx, doc *document.Document) (int64, error) { //NOSONAR
	return mockInt64Result(m.Called(tx, doc))
}

func (m *mockRepo) SaveMetadataWithObjId(tx *sql.Tx, objId *int64, doc *document.Document) (int64, error) {
	args := m.Called(tx, objId, doc)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockRepo) SaveMetadata(tx *sql.Tx, doc *document.Document) (int64, error) {
	return mockInt64Result(m.Called(tx, doc))
}

func (m *mockRepo) InsertFilePath(tx *sql.Tx, doc *document.Document) error {
	return m.Called(tx, doc).Error(0)
}

func (m *mockRepo) SaveItems(tx *sql.Tx, docId int64, items []document.Item) error {
	return m.Called(tx, docId, items).Error(0)
}

func (m *mockRepo) BeginTx() (*sql.Tx, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sql.Tx), args.Error(1)
}

func (m *mockRepo) GetPurchasesByUser(userID int, year, month string) ([]document.Document, error) {
	args := m.Called(userID, year, month)
	return args.Get(0).([]document.Document), args.Error(1)
}

func (m *mockRepo) GetPurchaseByObjId(objId int64, userID int) (*document.Document, error) {
	args := m.Called(objId, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*document.Document), args.Error(1)
}

func (m *mockRepo) GetItemsByPurchaseId(objId int64, userID int) ([]document.Item, error) {
	args := m.Called(objId, userID)
	return args.Get(0).([]document.Item), args.Error(1)
}

func (m *mockRepo) GetFilePathByObjId(objId int64, userID int) (string, string, error) {
	args := m.Called(objId, userID)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *mockRepo) GetDocIdByObjId(tx *sql.Tx, objId int64, userID int) (int64, error) {
	return mockInt64Result(m.Called(tx, objId, userID))
}

func (m *mockRepo) UpdatePurchaseMetadata(tx *sql.Tx, objId int64, userID int, doc *document.Document) error {
	return m.Called(tx, objId, userID, doc).Error(0)
}

func (m *mockRepo) DeleteItemsByDocId(tx *sql.Tx, docId int64) error {
	return m.Called(tx, docId).Error(0)
}

func (m *mockRepo) UpsertFilePath(tx *sql.Tx, doc *document.Document) error {
	return m.Called(tx, doc).Error(0)
}

func (m *mockRepo) SoftDeletePurchase(objId int64, userID int) error {
	return m.Called(objId, userID).Error(0)
}

const errDBMsg = "db error"

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

// TestSaveItemsWithItems verifies items are persisted under the correct objId and docId.
func TestSaveItemsWithItems(t *testing.T) {
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
	repo.On("SaveItems", mock.Anything, int64(10), items).Return(nil)

	svc := testutil.NewDocumentService(t, repo)
	err := svc.UploadDocument(&document.UploadInput{
		Title:  "Woolworths Receipt",
		UserId: 1,
		Items:  items,
	}, nil)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

// TestSaveItemsNilItems verifies SaveItems is still called when no items are provided.
func TestSaveItemsNilItems(t *testing.T) {
	db, dbMock, tx := setupTx(t)
	defer db.Close()
	dbMock.ExpectCommit()

	repo := new(mockRepo)
	repo.On("BeginTx").Return(tx, nil)
	repo.On("SetLatestObjId", mock.Anything, mock.Anything).Return(int64(1), nil)
	repo.On("SaveMetadataWithObjId", mock.Anything, mock.Anything, mock.Anything).Return(int64(2), nil)
	repo.On("SaveItems", mock.Anything, int64(2), mock.Anything).Return(nil)

	svc := testutil.NewDocumentService(t, repo)
	err := svc.UploadDocument(&document.UploadInput{
		Title:  "Empty Receipt",
		UserId: 1,
	}, nil)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

// TestSaveItemsError verifies that a SaveItems failure causes an error and rolls back.
func TestSaveItemsError(t *testing.T) {
	db, dbMock, tx := setupTx(t)
	defer db.Close()
	dbMock.ExpectRollback()

	repo := new(mockRepo)
	repo.On("BeginTx").Return(tx, nil)
	repo.On("SetLatestObjId", mock.Anything, mock.Anything).Return(int64(1), nil)
	repo.On("SaveMetadataWithObjId", mock.Anything, mock.Anything, mock.Anything).Return(int64(2), nil)
	repo.On("SaveItems", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("db insert failed"))

	svc := testutil.NewDocumentService(t, repo)
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

// TestSaveItemsMetadataError verifies that a metadata failure prevents SaveItems from being called.
func TestSaveItemsMetadataError(t *testing.T) {
	db, dbMock, tx := setupTx(t)
	defer db.Close()
	dbMock.ExpectRollback()

	repo := new(mockRepo)
	repo.On("BeginTx").Return(tx, nil)
	repo.On("SetLatestObjId", mock.Anything, mock.Anything).Return(int64(1), nil)
	repo.On("SaveMetadataWithObjId", mock.Anything, mock.Anything, mock.Anything).Return(int64(0), errors.New("metadata insert failed"))

	svc := testutil.NewDocumentService(t, repo)
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

// --- GetPurchases ---

// TestGetPurchasesSuccess verifies that GetPurchases returns the documents from the repository.
func TestGetPurchasesSuccess(t *testing.T) {
	expected := []document.Document{
		{Id: "1", Title: "Woolworths", Items: []document.Item{}},
		{Id: "2", Title: "Coles", Items: []document.Item{}},
	}

	repo := new(mockRepo)
	repo.On("GetPurchasesByUser", 1, "", "").Return(expected, nil)

	svc := testutil.NewDocumentService(t, repo)
	got, err := svc.GetPurchases(1, "", "")

	assert.NoError(t, err)
	assert.Equal(t, expected, got)
	repo.AssertExpectations(t)
}

// TestGetPurchasesWithFilters verifies that year and month are forwarded to the repository.
func TestGetPurchasesWithFilters(t *testing.T) {
	tests := []struct {
		name  string
		year  string
		month string
	}{
		{"year only", "2024", ""},
		{"month only", "", "03"},
		{"year and month", "2024", "03"},
		{"all wildcard", "all", "all"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockRepo)
			repo.On("GetPurchasesByUser", 1, tt.year, tt.month).Return([]document.Document{}, nil)

			svc := testutil.NewDocumentService(t, repo)
			_, err := svc.GetPurchases(1, tt.year, tt.month)

			assert.NoError(t, err)
			repo.AssertExpectations(t)
		})
	}
}

// TestGetPurchasesError verifies that a repository error is propagated.
func TestGetPurchasesError(t *testing.T) {
	repo := new(mockRepo)
	repo.On("GetPurchasesByUser", 1, "", "").Return([]document.Document{}, errors.New(errDBMsg))

	svc := testutil.NewDocumentService(t, repo)
	_, err := svc.GetPurchases(1, "", "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), errDBMsg)
	repo.AssertExpectations(t)
}

// --- GetPurchaseItems ---

// TestGetPurchaseItemsSuccess verifies that GetPurchaseItems returns the items from the repository.
func TestGetPurchaseItemsSuccess(t *testing.T) {
	expected := []document.Item{
		{Name: "Apple", Quantity: "2", UnitPrice: "1.50", SubTotal: "3.00"},
		{Name: "Bread", Quantity: "1", UnitPrice: "3.00", SubTotal: "3.00"},
	}

	repo := new(mockRepo)
	repo.On("GetItemsByPurchaseId", int64(5), 1).Return(expected, nil)

	svc := testutil.NewDocumentService(t, repo)
	got, err := svc.GetPurchaseItems(5, 1)

	assert.NoError(t, err)
	assert.Equal(t, expected, got)
	repo.AssertExpectations(t)
}

// TestGetPurchaseItemsEmpty verifies that an empty slice is returned when a purchase has no items.
func TestGetPurchaseItemsEmpty(t *testing.T) {
	repo := new(mockRepo)
	repo.On("GetItemsByPurchaseId", int64(5), 1).Return([]document.Item{}, nil)

	svc := testutil.NewDocumentService(t, repo)
	got, err := svc.GetPurchaseItems(5, 1)

	assert.NoError(t, err)
	assert.Empty(t, got)
	repo.AssertExpectations(t)
}

// TestGetPurchaseItemsError verifies that a repository error is propagated.
func TestGetPurchaseItemsError(t *testing.T) {
	repo := new(mockRepo)
	repo.On("GetItemsByPurchaseId", int64(99), 1).Return([]document.Item{}, errors.New(errDBMsg))

	svc := testutil.NewDocumentService(t, repo)
	_, err := svc.GetPurchaseItems(99, 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), errDBMsg)
	repo.AssertExpectations(t)
}

// --- GetFilePath ---

// TestGetFilePathSuccess verifies that file path and name are returned for a valid document.
func TestGetFilePathSuccess(t *testing.T) {
	repo := new(mockRepo)
	repo.On("GetFilePathByObjId", int64(5), 1).Return("./filedata/0/0/5.pdf", testReceiptFilename, nil)

	svc := testutil.NewDocumentService(t, repo)
	path, name, err := svc.GetFilePath(5, 1)

	assert.NoError(t, err)
	assert.Equal(t, "./filedata/0/0/5.pdf", path)
	assert.Equal(t, testReceiptFilename, name)
	repo.AssertExpectations(t)
}

// TestGetFilePathError verifies that a repository error is propagated.
func TestGetFilePathError(t *testing.T) {
	repo := new(mockRepo)
	repo.On("GetFilePathByObjId", int64(99), 1).Return("", "", errors.New("not found"))

	svc := testutil.NewDocumentService(t, repo)
	_, _, err := svc.GetFilePath(99, 1)

	assert.Error(t, err)
	repo.AssertExpectations(t)
}

// --- GetPurchaseById ---

// TestGetPurchaseByIdSuccess verifies that GetPurchaseById returns the document from the repository.
func TestGetPurchaseByIdSuccess(t *testing.T) {
	expected := &document.Document{
		Id:    "5",
		Title: "Woolworths",
		Items: []document.Item{
			{Name: "Apple", Quantity: "1", UnitPrice: "1.50", SubTotal: "1.50"},
		},
	}

	repo := new(mockRepo)
	repo.On("GetPurchaseByObjId", int64(5), 1).Return(expected, nil)

	svc := testutil.NewDocumentService(t, repo)
	got, err := svc.GetPurchaseById(5, 1)

	assert.NoError(t, err)
	assert.Equal(t, expected, got)
	repo.AssertExpectations(t)
}

// TestGetPurchaseByIdNotFound verifies that sql.ErrNoRows is returned when the purchase is not found.
func TestGetPurchaseByIdNotFound(t *testing.T) {
	repo := new(mockRepo)
	repo.On("GetPurchaseByObjId", int64(99), 1).Return(nil, sql.ErrNoRows)

	svc := testutil.NewDocumentService(t, repo)
	got, err := svc.GetPurchaseById(99, 1)

	assert.Nil(t, got)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, sql.ErrNoRows))
	repo.AssertExpectations(t)
}

// TestGetPurchaseByIdError verifies that a repository error is propagated.
func TestGetPurchaseByIdError(t *testing.T) {
	repo := new(mockRepo)
	repo.On("GetPurchaseByObjId", int64(5), 1).Return(nil, errors.New(errDBMsg))

	svc := testutil.NewDocumentService(t, repo)
	_, err := svc.GetPurchaseById(5, 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), errDBMsg)
	repo.AssertExpectations(t)
}

// --- UpdatePurchase ---

// TestUpdatePurchaseSuccess verifies that UpdatePurchase commits all steps successfully.
func TestUpdatePurchaseSuccess(t *testing.T) {
	items := []document.Item{
		{Name: "Milk", Quantity: "2", UnitPrice: "2.00", SubTotal: "4.00"},
	}

	db, dbMock, tx := setupTx(t)
	defer db.Close()
	dbMock.ExpectCommit()

	repo := new(mockRepo)
	repo.On("BeginTx").Return(tx, nil)
	repo.On("UpdatePurchaseMetadata", mock.Anything, int64(5), 1, mock.Anything).Return(nil)
	repo.On("GetDocIdByObjId", mock.Anything, int64(5), 1).Return(int64(10), nil)
	repo.On("DeleteItemsByDocId", mock.Anything, int64(10)).Return(nil)
	repo.On("SaveItems", mock.Anything, int64(10), items).Return(nil)

	svc := testutil.NewDocumentService(t, repo)
	err := svc.UpdatePurchase(5, 1, &document.UploadInput{
		Title:   "Updated Receipt",
		BuyFrom: "Coles",
		Items:   items,
		UserId:  1,
	})

	assert.NoError(t, err)
	repo.AssertExpectations(t)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

// TestUpdatePurchaseNotFound verifies that sql.ErrNoRows is propagated when the document is not found.
func TestUpdatePurchaseNotFound(t *testing.T) {
	db, dbMock, tx := setupTx(t)
	defer db.Close()
	dbMock.ExpectRollback()

	repo := new(mockRepo)
	repo.On("BeginTx").Return(tx, nil)
	repo.On("UpdatePurchaseMetadata", mock.Anything, int64(99), 1, mock.Anything).Return(sql.ErrNoRows)

	svc := testutil.NewDocumentService(t, repo)
	err := svc.UpdatePurchase(99, 1, &document.UploadInput{Title: "Ghost", UserId: 1})

	assert.Error(t, err)
	assert.True(t, errors.Is(err, sql.ErrNoRows))
	repo.AssertExpectations(t)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

// TestUpdatePurchaseDeleteItemsError verifies that a DeleteItems failure rolls back the transaction.
func TestUpdatePurchaseDeleteItemsError(t *testing.T) {
	db, dbMock, tx := setupTx(t)
	defer db.Close()
	dbMock.ExpectRollback()

	repo := new(mockRepo)
	repo.On("BeginTx").Return(tx, nil)
	repo.On("UpdatePurchaseMetadata", mock.Anything, int64(5), 1, mock.Anything).Return(nil)
	repo.On("GetDocIdByObjId", mock.Anything, int64(5), 1).Return(int64(10), nil)
	repo.On("DeleteItemsByDocId", mock.Anything, int64(10)).Return(errors.New("db delete failed"))

	svc := testutil.NewDocumentService(t, repo)
	err := svc.UpdatePurchase(5, 1, &document.UploadInput{Title: "Receipt", UserId: 1})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete items")
	repo.AssertExpectations(t)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

// --- DeletePurchase ---

// TestDeletePurchaseSuccess verifies that DeletePurchase delegates to SoftDeletePurchase.
func TestDeletePurchaseSuccess(t *testing.T) {
	repo := new(mockRepo)
	repo.On("SoftDeletePurchase", int64(5), 1).Return(nil)

	svc := testutil.NewDocumentService(t, repo)
	err := svc.DeletePurchase(5, 1)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

// TestDeletePurchaseNotFound verifies that sql.ErrNoRows is returned when the document is not found.
func TestDeletePurchaseNotFound(t *testing.T) {
	repo := new(mockRepo)
	repo.On("SoftDeletePurchase", int64(99), 1).Return(sql.ErrNoRows)

	svc := testutil.NewDocumentService(t, repo)
	err := svc.DeletePurchase(99, 1)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, sql.ErrNoRows))
	repo.AssertExpectations(t)
}

// --- UpdatePurchaseWithFile ---

const testReceiptFilename = "receipt.pdf"

// minimalFileHeader returns a *multipart.FileHeader with only the fields
// read during Document construction (Filename, Size). The underlying file
// is never opened in the error-path tests below.
func minimalFileHeader(name string) *multipart.FileHeader {
	return &multipart.FileHeader{Filename: name, Size: 0}
}

// TestUpdatePurchaseWithFileNotFound verifies that sql.ErrNoRows is propagated
// when the document does not belong to the user.
func TestUpdatePurchaseWithFileNotFound(t *testing.T) {
	db, dbMock, tx := setupTx(t)
	defer db.Close()
	dbMock.ExpectRollback()

	repo := new(mockRepo)
	repo.On("BeginTx").Return(tx, nil)
	repo.On("UpdatePurchaseMetadata", mock.Anything, int64(99), 1, mock.Anything).Return(sql.ErrNoRows)

	svc := testutil.NewDocumentService(t, repo)
	err := svc.UpdatePurchaseWithFile(99, 1, &document.UploadInput{
		Title:  "Ghost",
		UserId: 1,
		File:   minimalFileHeader(testReceiptFilename),
	}, nil)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, sql.ErrNoRows))
	repo.AssertExpectations(t)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

// TestUpdatePurchaseWithFileDeleteItemsError verifies that a DeleteItems failure
// rolls back the transaction and returns an error.
func TestUpdatePurchaseWithFileDeleteItemsError(t *testing.T) {
	db, dbMock, tx := setupTx(t)
	defer db.Close()
	dbMock.ExpectRollback()

	repo := new(mockRepo)
	repo.On("BeginTx").Return(tx, nil)
	repo.On("UpdatePurchaseMetadata", mock.Anything, int64(5), 1, mock.Anything).Return(nil)
	repo.On("GetDocIdByObjId", mock.Anything, int64(5), 1).Return(int64(10), nil)
	repo.On("DeleteItemsByDocId", mock.Anything, int64(10)).Return(errors.New("db delete failed"))

	svc := testutil.NewDocumentService(t, repo)
	err := svc.UpdatePurchaseWithFile(5, 1, &document.UploadInput{
		Title:  "Receipt",
		UserId: 1,
		File:   minimalFileHeader(testReceiptFilename),
	}, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete items")
	repo.AssertExpectations(t)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

// TestUpdatePurchaseWithFileSaveItemsError verifies that a SaveItems failure
// rolls back the transaction and returns an error.
func TestUpdatePurchaseWithFileSaveItemsError(t *testing.T) {
	db, dbMock, tx := setupTx(t)
	defer db.Close()
	dbMock.ExpectRollback()

	repo := new(mockRepo)
	repo.On("BeginTx").Return(tx, nil)
	repo.On("UpdatePurchaseMetadata", mock.Anything, int64(5), 1, mock.Anything).Return(nil)
	repo.On("GetDocIdByObjId", mock.Anything, int64(5), 1).Return(int64(10), nil)
	repo.On("DeleteItemsByDocId", mock.Anything, int64(10)).Return(nil)
	repo.On("SaveItems", mock.Anything, int64(10), mock.Anything).Return(errors.New("db insert failed"))

	svc := testutil.NewDocumentService(t, repo)
	err := svc.UpdatePurchaseWithFile(5, 1, &document.UploadInput{
		Title:  "Receipt",
		UserId: 1,
		File:   minimalFileHeader(testReceiptFilename),
	}, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "save items")
	repo.AssertExpectations(t)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}
