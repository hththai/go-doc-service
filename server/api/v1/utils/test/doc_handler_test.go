package test

import (
	"2_Go/api/v1/utils"
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

func (m *MockDocumentRepository) SaveMetadata(tx *sql.Tx, doc *document.Document) (int64, error) {
	args := m.Called(tx, doc)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockDocumentRepository) InsertFilePath(tx *sql.Tx, doc *document.Document) error {
	args := m.Called(tx, doc)
	return args.Error(0)
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
		setUserId     bool
		mockRepo      func(repo *MockDocumentRepository)
		mockDB        func() (*sql.DB, sqlmock.Sqlmock)
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
			setUserId:   true,
			mockRepo: func(repo *MockDocumentRepository) {
				repo.On("SetLatestObjId", mock.Anything, mock.Anything).Return(int64(1), nil)
				repo.On("SaveMetadataWithObjId", mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
				repo.On("InsertFilePath", mock.Anything, mock.Anything).Return(nil)
			},
			mockDB: func() (*sql.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				mock.ExpectBegin()
				mock.ExpectCommit()
				return db, mock
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
			setUserId:   true,
			mockRepo: func(repo *MockDocumentRepository) {
				repo.On("SetLatestObjId", mock.Anything, mock.Anything).Return(int64(2), nil)
				repo.On("SaveMetadataWithObjId", mock.Anything, mock.Anything, mock.Anything).Return(int64(2), nil)
			},
			mockDB: func() (*sql.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				mock.ExpectBegin()
				mock.ExpectCommit()
				return db, mock
			},
			expectErr: false,
		},
		{
			name: "error - user not authenticated",
			formData: map[string]string{
				"name":        "testDocumentName",
				"description": "testDescription",
			},
			includeFile: false,
			setUserId:   false,
			mockRepo: func(repo *MockDocumentRepository) {
				// No mock needed as it should fail before repository call
			},
			mockDB: func() (*sql.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				return db, mock
			},
			expectErr:     true,
			errorContains: "user not authenticated",
		},
		{
			name: "error - transaction begin fails",
			formData: map[string]string{
				"name":        "testDocumentName",
				"description": "testDescription",
			},
			includeFile: false,
			userId:      123,
			setUserId:   true,
			mockRepo: func(repo *MockDocumentRepository) {
				// No mock needed as transaction should fail before repository call
			},
			mockDB: func() (*sql.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				mock.ExpectBegin().WillReturnError(errors.New("failed to start transaction"))
				return db, mock
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
			setUserId:   true,
			mockRepo: func(repo *MockDocumentRepository) {
				repo.On("SetLatestObjId", mock.Anything, mock.Anything).
					Return(int64(-1), errors.New("failed to set obj id"))
			},
			mockDB: func() (*sql.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				mock.ExpectBegin()
				mock.ExpectRollback()
				return db, mock
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
			setUserId:   true,
			mockRepo: func(repo *MockDocumentRepository) {
				repo.On("SetLatestObjId", mock.Anything, mock.Anything).Return(int64(3), nil)
				repo.On("SaveMetadataWithObjId", mock.Anything, mock.Anything, mock.Anything).
					Return(int64(-1), errors.New("failed to save metadata"))
			},
			mockDB: func() (*sql.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				mock.ExpectBegin()
				mock.ExpectRollback()
				return db, mock
			},
			expectErr:     true,
			errorContains: "metadata save failed",
		},
		{
			name: "error - commit fails",
			formData: map[string]string{
				"name":        "testDocumentName",
				"description": "testDescription",
			},
			includeFile: false,
			userId:      123,
			setUserId:   true,
			mockRepo: func(repo *MockDocumentRepository) {
				repo.On("SetLatestObjId", mock.Anything, mock.Anything).Return(int64(4), nil)
				repo.On("SaveMetadataWithObjId", mock.Anything, mock.Anything, mock.Anything).Return(int64(4), nil)
			},
			mockDB: func() (*sql.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				mock.ExpectBegin()
				mock.ExpectCommit().WillReturnError(errors.New("commit failed"))
				return db, mock
			},
			expectErr:     true,
			errorContains: "commit failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Create multipart form
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			// Add form fields
			for key, val := range tt.formData {
				_ = writer.WriteField(key, val)
			}

			// Add file if needed
			if tt.includeFile {
				part, _ := writer.CreateFormFile("file", tt.fileName)
				_, _ = part.Write([]byte(tt.fileContent))
			}

			_ = writer.Close()

			// Create request
			c.Request, _ = http.NewRequest("POST", "/upload", body)
			c.Request.Header.Set("Content-Type", writer.FormDataContentType())

			// Set user ID in context if needed
			if tt.setUserId {
				c.Set("userId", tt.userId)
			}

			// Setup mocks
			mockRepo := new(MockDocumentRepository)
			tt.mockRepo(mockRepo)

			db, mockDB := tt.mockDB()
			defer db.Close()

			service := document.NewDocumentService(mockRepo)

			// Execute
			err := utils.UploadDocument(c, service, db)

			// Assertions
			if tt.expectErr {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			// Verify mock expectations
			mockRepo.AssertExpectations(t)
			assert.NoError(t, mockDB.ExpectationsWereMet())
		})
	}
}
