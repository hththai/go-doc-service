package documentApi_test

import (
	"2_Go/api/v1/documentApi"
	"2_Go/internal/document"
	"bytes"
	"database/sql"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDocumentRepository mocks the document repository interface.
type MockDocumentRepository struct {
	mock.Mock
}

// mockInt64Result extracts (int64, error) from a testify mock call result.
func mockInt64Result(args mock.Arguments) (int64, error) {
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockDocumentRepository) SetLatestObjId(tx *sql.Tx, doc *document.Document) (int64, error) {
	return mockInt64Result(m.Called(tx, doc))
}

func (m *MockDocumentRepository) SaveMetadataWithObjId(tx *sql.Tx, objId *int64, doc *document.Document) (int64, error) {
	return mockInt64Result(m.Called(tx, objId, doc))
}

func (m *MockDocumentRepository) SaveMetadata(tx *sql.Tx, doc *document.Document) (int64, error) {
	return mockInt64Result(m.Called(tx, doc))
}

func (m *MockDocumentRepository) InsertFilePath(tx *sql.Tx, doc *document.Document) error {
	args := m.Called(tx, doc)
	return args.Error(0)
}

func (m *MockDocumentRepository) BeginTx() (*sql.Tx, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sql.Tx), args.Error(1)
}

// buildForm creates a multipart form body from the given fields.
func buildForm(fields map[string]string) (*bytes.Buffer, string) {
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	for k, v := range fields {
		_ = w.WriteField(k, v)
	}
	_ = w.Close()
	return body, w.FormDataContentType()
}

func TestHandleUploadPurchaseInfoBuyAt(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		buyAt          string
		expectedStatus int
		// checkBuyAt is called with the BuyAt stored in the captured Document (nil for error cases).
		checkBuyAt func(t *testing.T, buyAt *time.Time)
	}{
		{
			name:           "valid Australian date DD/MM/YYYY",
			buyAt:          "22/02/2026",
			expectedStatus: http.StatusOK,
			checkBuyAt: func(t *testing.T, buyAt *time.Time) {
				assert.NotNil(t, buyAt)
				assert.Equal(t, 22, buyAt.Day())
				assert.Equal(t, time.February, buyAt.Month())
				assert.Equal(t, 2026, buyAt.Year())
			},
		},
		{
			name:           "empty buyAt is optional (nil)",
			buyAt:          "",
			expectedStatus: http.StatusOK,
			checkBuyAt: func(t *testing.T, buyAt *time.Time) {
				assert.Nil(t, buyAt)
			},
		},
		{
			name:           "invalid format YYYY-MM-DD returns 400",
			buyAt:          "2026-02-22",
			expectedStatus: http.StatusBadRequest,
			checkBuyAt:     nil,
		},
		{
			name:           "American format MM/DD/YYYY returns 400",
			buyAt:          "02/22/2026",
			expectedStatus: http.StatusBadRequest,
			checkBuyAt:     nil,
		},
		{
			name:           "garbage value returns 400",
			buyAt:          "not-a-date",
			expectedStatus: http.StatusBadRequest,
			checkBuyAt:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := map[string]string{
				"name":        "Test Receipt",
				"description": "OCR purchase",
				"buyFrom":     "Woolworths",
				"buyPrice":    "12.50",
			}
			if tt.buyAt != "" {
				fields["buyAt"] = tt.buyAt
			}

			body, contentType := buildForm(fields)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest("POST", "/upload", body)
			c.Request.Header.Set("Content-Type", contentType)
			c.Set("userId", 1)

			mockRepo := new(MockDocumentRepository)
			logger, _ := test.NewNullLogger()

			// For success cases, wire up mock repo and capture the document.
			var capturedDoc *document.Document
			if tt.checkBuyAt != nil {
				db, mockDB, err := sqlmock.New()
				assert.NoError(t, err)
				defer db.Close()

				mockDB.ExpectBegin()
				mockDB.ExpectCommit()
				tx, _ := db.Begin()

				mockRepo.On("BeginTx").Return(tx, nil)
				mockRepo.On("SetLatestObjId", mock.Anything, mock.Anything).Return(int64(1), nil)
				mockRepo.On("SaveMetadataWithObjId", mock.Anything, mock.Anything,
					mock.MatchedBy(func(doc *document.Document) bool {
						capturedDoc = doc
						return true
					}),
				).Return(int64(1), nil)

				defer func() {
					assert.NoError(t, mockDB.ExpectationsWereMet())
				}()
			}

			svc := document.NewDocumentService(mockRepo)
			h := &documentApi.DocumentHandler{DocSvc: *svc, Logger: logger}
			h.HandleUpload(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockRepo.AssertExpectations(t)

			if tt.checkBuyAt != nil {
				assert.NotNil(t, capturedDoc, "expected service to be called with a document")
				tt.checkBuyAt(t, capturedDoc.PurchaseInfo.BuyAt)
			}
		})
	}
}
