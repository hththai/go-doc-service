package test

import (
	"2_Go/internal/document"
	"bytes"
	"database/sql"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testDocumentName = "Test Document"
	testDescription  = "Test Description"
)

// MockDocumentRepository mocks the document repository
type MockDocumentRepository struct {
	mock.Mock
}

func (m *MockDocumentRepository) SetLatestObjId(tx *sql.Tx, doc *document.Document) (int64, error) {
	args := m.Called(tx, doc)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockDocumentRepository) SaveMetadataWithObjId(tx *sql.Tx, objId *int64, doc *document.Document) (int64, error) {
	args := m.Called(tx, objId, doc)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockDocumentRepository) SaveMetadata(tx *sql.Tx, doc *document.Document) (id int64, err error) {
	args := m.Called(tx, doc)
	id = args.Get(0).(int64)
	err = args.Error(1)
	return
}

func (m *MockDocumentRepository) InsertFilePath(tx *sql.Tx, doc *document.Document) error {
	args := m.Called(tx, doc)
	return args.Error(0)
}

func (m *MockDocumentRepository) SaveItems(tx *sql.Tx, objId int64, docId int64, items []document.Item) error {
	return m.Called(tx, objId, docId, items).Error(0)
}

func (m *MockDocumentRepository) BeginTx() (*sql.Tx, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sql.Tx), args.Error(1)
}

func (m *MockDocumentRepository) GetPurchasesByUser(userID int, year, month string) ([]document.Document, error) {
	args := m.Called(userID, year, month)
	return args.Get(0).([]document.Document), args.Error(1)
}

func (m *MockDocumentRepository) GetFilePathByObjId(objId int64, userID int) (string, string, error) {
	args := m.Called(objId, userID)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockDocumentRepository) GetDocIdByObjId(tx *sql.Tx, objId int64, userID int) (int64, error) {
	args := m.Called(tx, objId, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockDocumentRepository) UpdatePurchaseMetadata(tx *sql.Tx, objId int64, userID int, doc *document.Document) error {
	return m.Called(tx, objId, userID, doc).Error(0)
}

func (m *MockDocumentRepository) DeleteItemsByObjId(tx *sql.Tx, objId int64) error {
	return m.Called(tx, objId).Error(0)
}

func (m *MockDocumentRepository) UpsertFilePath(tx *sql.Tx, doc *document.Document) error {
	return m.Called(tx, doc).Error(0)
}

func (m *MockDocumentRepository) SoftDeletePurchase(objId int64, userID int) error {
	return m.Called(objId, userID).Error(0)
}

// Helper function to create multipart request
func createMultipartRequest(formData map[string]string, includeFile bool, fileName, fileContent string) (*bytes.Buffer, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for key, val := range formData {
		_ = writer.WriteField(key, val)
	}

	if includeFile {
		part, _ := writer.CreateFormFile("file", fileName)
		_, _ = part.Write([]byte(fileContent))
	}

	_ = writer.Close()
	return body, writer.FormDataContentType()
}

// Helper function to setup test context
func setupTestContext(body *bytes.Buffer, contentType string, setUserId bool, userId int) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request, _ = http.NewRequest("POST", "/upload", body)
	c.Request.Header.Set("Content-Type", contentType)

	if setUserId {
		c.Set("userId", userId)
	}

	return c, w
}

// Helper function to verify test results
func verifyTestResult(t *testing.T, err error, expectErr bool, errorContains string, mockRepo *MockDocumentRepository, mockDB sqlmock.Sqlmock) {
	if expectErr {
		assert.Error(t, err)
		if errorContains != "" {
			assert.Contains(t, err.Error(), errorContains)
		}
	} else {
		assert.NoError(t, err)
	}

	mockRepo.AssertExpectations(t)
	assert.NoError(t, mockDB.ExpectationsWereMet())
}

// mockSaveFile is a mock file save function for testing
func mockSaveFile(file *multipart.FileHeader, dst string) error {
	return nil
}

func TestUploadDocument(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		formData      map[string]string
		includeFile   bool
		fileName      string
		fileContent   string
		userId        int
		mockRepo      func(repo *MockDocumentRepository, db *sql.DB, mockDB sqlmock.Sqlmock)
		expectErr     bool
		errorContains string
	}{
		{
			name: "success - upload with file",
			formData: map[string]string{
				"name":        testDocumentName,
				"description": testDescription,
			},
			includeFile: true,
			fileName:    "test.pdf",
			fileContent: "dummy file content",
			userId:      123,
			mockRepo: func(repo *MockDocumentRepository, db *sql.DB, mockDB sqlmock.Sqlmock) {
				mockDB.ExpectBegin()
				mockDB.ExpectCommit()
				tx, _ := db.Begin()
				repo.On("BeginTx").Return(tx, nil)
				repo.On("SetLatestObjId", mock.Anything, mock.Anything).Return(int64(1), nil)
				repo.On("SaveMetadataWithObjId", mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
				repo.On("InsertFilePath", mock.Anything, mock.Anything).Return(nil)
				repo.On("SaveItems", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectErr: false,
		},
		{
			name: "success - upload without file (metadata only)",
			formData: map[string]string{
				"name":        "Metadata Only",
				"description": "No file attached",
			},
			includeFile: false,
			userId:      456,
			mockRepo: func(repo *MockDocumentRepository, db *sql.DB, mockDB sqlmock.Sqlmock) {
				mockDB.ExpectBegin()
				mockDB.ExpectCommit()
				tx, _ := db.Begin()
				repo.On("BeginTx").Return(tx, nil)
				repo.On("SetLatestObjId", mock.Anything, mock.Anything).Return(int64(2), nil)
				repo.On("SaveMetadataWithObjId", mock.Anything, mock.Anything, mock.Anything).Return(int64(2), nil)
				repo.On("SaveItems", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectErr: false,
		},
		{
			name: "error - transaction begin fails",
			formData: map[string]string{
				"name":        "testDocumentName",
				"description": "testDescription",
			},
			includeFile: false,
			userId:      123,
			mockRepo: func(repo *MockDocumentRepository, db *sql.DB, mockDB sqlmock.Sqlmock) {
				repo.On("BeginTx").Return(nil, errors.New("failed to start transaction"))
			},
			expectErr:     true,
			errorContains: "failed to start transaction",
		},
		{
			name: "error - SetLatestObjID fails",
			formData: map[string]string{
				"name":        "testDocumentName",
				"description": "testDescription",
			},
			includeFile: false,
			userId:      123,
			mockRepo: func(repo *MockDocumentRepository, db *sql.DB, mockDB sqlmock.Sqlmock) {
				mockDB.ExpectBegin()
				mockDB.ExpectRollback()
				tx, _ := db.Begin()
				repo.On("BeginTx").Return(tx, nil)
				repo.On("SetLatestObjId", mock.Anything, mock.Anything).
					Return(int64(-1), errors.New("failed to set obj id"))
			},
			expectErr:     true,
			errorContains: "metadata save failed",
		},
		{
			name: "error - SaveMetadataWithObjId fails (no file)",
			formData: map[string]string{
				"name":        "testDocumentName",
				"description": "testDescription",
			},
			includeFile: false,
			userId:      123,
			mockRepo: func(repo *MockDocumentRepository, db *sql.DB, mockDB sqlmock.Sqlmock) {
				mockDB.ExpectBegin()
				mockDB.ExpectRollback()
				tx, _ := db.Begin()
				repo.On("BeginTx").Return(tx, nil)
				repo.On("SetLatestObjId", mock.Anything, mock.Anything).Return(int64(3), nil)
				repo.On("SaveMetadataWithObjId", mock.Anything, mock.Anything, mock.Anything).
					Return(int64(-1), errors.New("failed to save metadata"))
			},
			expectErr:     true,
			errorContains: "failed to save metadata",
		},
		{
			name: "error - SaveItems fails",
			formData: map[string]string{
				"name":        "testDocumentName",
				"description": "testDescription",
			},
			includeFile: false,
			userId:      123,
			mockRepo: func(repo *MockDocumentRepository, db *sql.DB, mockDB sqlmock.Sqlmock) {
				mockDB.ExpectBegin()
				mockDB.ExpectRollback()
				tx, _ := db.Begin()
				repo.On("BeginTx").Return(tx, nil)
				repo.On("SetLatestObjId", mock.Anything, mock.Anything).Return(int64(5), nil)
				repo.On("SaveMetadataWithObjId", mock.Anything, mock.Anything, mock.Anything).Return(int64(5), nil)
				repo.On("SaveItems", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(errors.New("failed to insert items"))
			},
			expectErr:     true,
			errorContains: "failed to save items",
		},
		{
			name: "error - commit fails",
			formData: map[string]string{
				"name":        "testDocumentName",
				"description": "testDescription",
			},
			includeFile: false,
			userId:      123,
			mockRepo: func(repo *MockDocumentRepository, db *sql.DB, mockDB sqlmock.Sqlmock) {
				mockDB.ExpectBegin()
				mockDB.ExpectCommit().WillReturnError(errors.New("commit failed"))
				tx, _ := db.Begin()
				repo.On("BeginTx").Return(tx, nil)
				repo.On("SetLatestObjId", mock.Anything, mock.Anything).Return(int64(4), nil)
				repo.On("SaveMetadataWithObjId", mock.Anything, mock.Anything, mock.Anything).Return(int64(4), nil)
				repo.On("SaveItems", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectErr:     true,
			errorContains: "commit failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create multipart request
			body, contentType := createMultipartRequest(tt.formData, tt.includeFile, tt.fileName, tt.fileContent)

			// Setup test context to parse multipart form
			c, _ := setupTestContext(body, contentType, true, tt.userId)

			// Setup mocks
			mockRepo := new(MockDocumentRepository)
			db, mockDB, _ := sqlmock.New()
			defer db.Close()

			tt.mockRepo(mockRepo, db, mockDB)

			service := document.NewDocumentService(mockRepo)

			// Build input from context
			input := &document.UploadInput{
				Title:       tt.formData["name"],
				Description: tt.formData["description"],
				UserId:      tt.userId,
			}

			// Get file if included
			if tt.includeFile {
				file, _ := c.FormFile("file")
				input.File = file
			}

			// Execute using service method
			err := service.UploadDocument(input, mockSaveFile)

			// Verify results
			verifyTestResult(t, err, tt.expectErr, tt.errorContains, mockRepo, mockDB)
		})
	}
}
