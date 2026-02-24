package documentApi_test

import (
	"2_Go/api/v1/documentApi"
	"2_Go/internal/document"
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
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
		checkBuyAt func(t *testing.T, buyAt *document.Date)
	}{
		{
			name:           "valid Australian date DD/MM/YYYY",
			buyAt:          "22/02/2026",
			expectedStatus: http.StatusOK,
			checkBuyAt: func(t *testing.T, buyAt *document.Date) {
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
			checkBuyAt: func(t *testing.T, buyAt *document.Date) {
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
				mockRepo.On("SaveItems", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

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

// --- HandleGetPurchases ---

// newGetPurchasesContext builds a test gin.Context for GET /purchases with optional query params.
func newGetPurchasesContext(userID int, query string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/purchases?"+query, nil)
	if userID != 0 {
		c.Set("userId", userID)
	}
	return c, w
}

// TestHandleGetPurchasesSuccess verifies that a list of purchases is returned as JSON.
func TestHandleGetPurchasesSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	buyAt := document.Date{Time: time.Date(2024, 3, 5, 0, 0, 0, 0, time.UTC)}
	docs := []document.Document{
		{
			Id:       "1",
			Title:    "Monthly Groceries",
			FileName: "coles_mar2024.pdf",
			FilePath: "./filedata/0/0/1.pdf",
			PurchaseInfo: document.PurchaseInfo{
				BuyAt:    &buyAt,
				BuyFrom:  "Coles",
				BuyPrice: "112.30",
			},
			Items: []document.Item{
				{Name: "Apple", Quantity: "2", UnitPrice: "1.50", SubTotal: "3.00"},
			},
		},
	}

	mockRepo := new(MockDocumentRepository)
	mockRepo.On("GetPurchasesByUser", 1, "", "").Return(docs, nil)

	logger, _ := test.NewNullLogger()
	svc := document.NewDocumentService(mockRepo)
	h := &documentApi.DocumentHandler{DocSvc: *svc, Logger: logger}

	c, w := newGetPurchasesContext(1, "")
	h.HandleGetPurchases(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp []documentApi.PurchaseResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Len(t, resp, 1)
	assert.Equal(t, "1", resp[0].ID)
	assert.Equal(t, "Monthly Groceries", resp[0].Title)
	assert.Equal(t, "2024-03-05", resp[0].BuyAt)
	assert.Equal(t, "Coles", resp[0].BuyFrom)
	assert.Equal(t, "112.30", resp[0].BuyPrice)
	assert.Equal(t, "/v1/auth/file/1", resp[0].FileURL)
	assert.Len(t, resp[0].Items, 1)
	assert.Equal(t, "Apple", resp[0].Items[0].ItemName)
	mockRepo.AssertExpectations(t)
}

// TestHandleGetPurchasesQueryParams verifies that year and month query params are forwarded.
func TestHandleGetPurchasesQueryParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name  string
		query string
		year  string
		month string
	}{
		{"no params", "", "", ""},
		{"year only", "year=2024", "2024", ""},
		{"month only", "month=03", "", "03"},
		{"year and month", "year=2024&month=03", "2024", "03"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockDocumentRepository)
			mockRepo.On("GetPurchasesByUser", 1, tt.year, tt.month).Return([]document.Document{}, nil)

			logger, _ := test.NewNullLogger()
			svc := document.NewDocumentService(mockRepo)
			h := &documentApi.DocumentHandler{DocSvc: *svc, Logger: logger}

			c, w := newGetPurchasesContext(1, tt.query)
			h.HandleGetPurchases(c)

			assert.Equal(t, http.StatusOK, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

// TestHandleGetPurchasesUnauthorized verifies that missing userId returns 401.
func TestHandleGetPurchasesUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger, _ := test.NewNullLogger()
	svc := document.NewDocumentService(new(MockDocumentRepository))
	h := &documentApi.DocumentHandler{DocSvc: *svc, Logger: logger}

	c, w := newGetPurchasesContext(0, "") // 0 = no userId set
	h.HandleGetPurchases(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestHandleGetPurchasesRepoError verifies that a repository error returns 500.
func TestHandleGetPurchasesRepoError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockDocumentRepository)
	mockRepo.On("GetPurchasesByUser", 1, "", "").Return([]document.Document{}, errors.New("db failure"))

	logger, _ := test.NewNullLogger()
	svc := document.NewDocumentService(mockRepo)
	h := &documentApi.DocumentHandler{DocSvc: *svc, Logger: logger}

	c, w := newGetPurchasesContext(1, "")
	h.HandleGetPurchases(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockRepo.AssertExpectations(t)
}

// TestHandleGetPurchasesNoFileURL verifies that fileUrl is omitted when no file is attached.
func TestHandleGetPurchasesNoFileURL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	docs := []document.Document{
		{Id: "2", Title: "No File", FilePath: "", PurchaseInfo: document.PurchaseInfo{}, Items: []document.Item{}},
	}
	mockRepo := new(MockDocumentRepository)
	mockRepo.On("GetPurchasesByUser", 1, "", "").Return(docs, nil)

	logger, _ := test.NewNullLogger()
	svc := document.NewDocumentService(mockRepo)
	h := &documentApi.DocumentHandler{DocSvc: *svc, Logger: logger}

	c, w := newGetPurchasesContext(1, "")
	h.HandleGetPurchases(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp []documentApi.PurchaseResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Len(t, resp, 1)
	assert.Empty(t, resp[0].FileURL)
}

// --- HandleServeFile ---

// TestHandleServeFileSuccess verifies that a valid file is served with 200.
func TestHandleServeFileSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Write a temp file to serve.
	f, err := os.CreateTemp(t.TempDir(), "receipt-*.pdf")
	assert.NoError(t, err)
	_, _ = f.WriteString("%PDF-1.4 fake content")
	f.Close()

	mockRepo := new(MockDocumentRepository)
	mockRepo.On("GetFilePathByObjId", int64(5), 1).Return(f.Name(), "receipt.pdf", nil)

	logger, _ := test.NewNullLogger()
	svc := document.NewDocumentService(mockRepo)
	h := &documentApi.DocumentHandler{DocSvc: *svc, Logger: logger}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/file/5", nil)
	c.Set("userId", 1)
	c.Params = gin.Params{{Key: "id", Value: "5"}}

	h.HandleServeFile(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Disposition"), "receipt.pdf")
	mockRepo.AssertExpectations(t)
}

// TestHandleServeFileNotFound verifies that a missing file path returns 404.
func TestHandleServeFileNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockDocumentRepository)
	mockRepo.On("GetFilePathByObjId", int64(99), 1).Return("", "", sql.ErrNoRows)

	logger, _ := test.NewNullLogger()
	svc := document.NewDocumentService(mockRepo)
	h := &documentApi.DocumentHandler{DocSvc: *svc, Logger: logger}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/file/99", nil)
	c.Set("userId", 1)
	c.Params = gin.Params{{Key: "id", Value: "99"}}

	h.HandleServeFile(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockRepo.AssertExpectations(t)
}

// TestHandleServeFileInvalidID verifies that a non-numeric id returns 400.
func TestHandleServeFileInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger, _ := test.NewNullLogger()
	svc := document.NewDocumentService(new(MockDocumentRepository))
	h := &documentApi.DocumentHandler{DocSvc: *svc, Logger: logger}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/file/abc", nil)
	c.Set("userId", 1)
	c.Params = gin.Params{{Key: "id", Value: "abc"}}

	h.HandleServeFile(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestHandleServeFileUnauthorized verifies that a missing userId returns 401.
func TestHandleServeFileUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger, _ := test.NewNullLogger()
	svc := document.NewDocumentService(new(MockDocumentRepository))
	h := &documentApi.DocumentHandler{DocSvc: *svc, Logger: logger}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/file/5", nil)
	c.Params = gin.Params{{Key: "id", Value: "5"}}
	// userId intentionally not set

	h.HandleServeFile(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
