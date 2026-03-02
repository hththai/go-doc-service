package documentApi_test

import (
	"2_Go/api/v1/documentApi"
	"2_Go/internal/document"
	"2_Go/internal/testutil"
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

func (m *MockDocumentRepository) UpsertFilePath(tx *sql.Tx, doc *document.Document) error {
	return m.Called(tx, doc).Error(0)
}

func (m *MockDocumentRepository) SaveItems(tx *sql.Tx, docId int64, items []document.Item) error {
	return m.Called(tx, docId, items).Error(0)
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

func (m *MockDocumentRepository) GetPurchaseByObjId(objId int64, userID int) (*document.Document, error) {
	args := m.Called(objId, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*document.Document), args.Error(1)
}

func (m *MockDocumentRepository) GetItemsByPurchaseId(objId int64, userID int) ([]document.Item, error) {
	args := m.Called(objId, userID)
	return args.Get(0).([]document.Item), args.Error(1)
}

func (m *MockDocumentRepository) GetFilePathByObjId(objId int64, userID int) (string, string, error) {
	args := m.Called(objId, userID)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockDocumentRepository) GetDocIdByObjId(tx *sql.Tx, objId int64, userID int) (int64, error) {
	return mockInt64Result(m.Called(tx, objId, userID))
}

func (m *MockDocumentRepository) UpdatePurchaseMetadata(tx *sql.Tx, objId int64, userID int, doc *document.Document) error {
	return m.Called(tx, objId, userID, doc).Error(0)
}

func (m *MockDocumentRepository) DeleteItemsByDocId(tx *sql.Tx, docId int64) error {
	return m.Called(tx, docId).Error(0)
}

func (m *MockDocumentRepository) SoftDeletePurchase(objId int64, userID int) error {
	return m.Called(objId, userID).Error(0)
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
			c.Request.Header.Set(headerContentType, contentType)
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
				mockRepo.On("SaveItems", mock.Anything, mock.Anything, mock.Anything).Return(nil)

				defer func() {
					assert.NoError(t, mockDB.ExpectationsWereMet())
				}()
			}

			svc := testutil.NewDocumentService(t, mockRepo)
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
	svc := testutil.NewDocumentService(t, mockRepo)
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
			svc := testutil.NewDocumentService(t, mockRepo)
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
	svc := testutil.NewDocumentService(t, new(MockDocumentRepository))
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
	svc := testutil.NewDocumentService(t, mockRepo)
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
	svc := testutil.NewDocumentService(t, mockRepo)
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
	svc := testutil.NewDocumentService(t, mockRepo)
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
	svc := testutil.NewDocumentService(t, mockRepo)
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
	svc := testutil.NewDocumentService(t, new(MockDocumentRepository))
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
	svc := testutil.NewDocumentService(t, new(MockDocumentRepository))
	h := &documentApi.DocumentHandler{DocSvc: *svc, Logger: logger}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/file/5", nil)
	c.Params = gin.Params{{Key: "id", Value: "5"}}
	// userId intentionally not set

	h.HandleServeFile(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// --- Helpers for JSON endpoints ---

const (
	pathPurchases    = "/purchases"
	pathPurchases5   = "/purchases/5"
	pathPurchases99  = "/purchases/99"
	pathPurchasesAbc = "/purchases/abc"

	headerContentType  = "Content-Type"
	testReceiptPDFName = "receipt.pdf"
)

// buildJSONContext creates a gin.Context with a JSON body for POST/PATCH/DELETE tests.
func buildJSONContext(method, path string, body interface{}, userID int, params gin.Params) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	bodyBytes, _ := json.Marshal(body)
	c.Request, _ = http.NewRequest(method, path, bytes.NewReader(bodyBytes))
	c.Request.Header.Set(headerContentType, "application/json")
	if userID != 0 {
		c.Set("userId", userID)
	}
	c.Params = params
	return c, w
}

// setupUpdateTx creates a sqlmock TX and wires the full UpdatePurchase mock chain.
func setupUpdateTx(t *testing.T, repo *MockDocumentRepository, objId int64, userID int, docId int64) sqlmock.Sqlmock {
	t.Helper()
	db, dbMock, err := sqlmock.New()
	assert.NoError(t, err)
	dbMock.ExpectBegin()
	dbMock.ExpectCommit()
	tx, _ := db.Begin()
	t.Cleanup(func() { db.Close() })

	repo.On("BeginTx").Return(tx, nil)
	repo.On("UpdatePurchaseMetadata", mock.Anything, objId, userID, mock.Anything).Return(nil)
	repo.On("GetDocIdByObjId", mock.Anything, objId, userID).Return(docId, nil)
	repo.On("DeleteItemsByDocId", mock.Anything, docId).Return(nil)
	repo.On("SaveItems", mock.Anything, docId, mock.Anything).Return(nil)
	return dbMock
}

// validPurchaseBody returns a minimal valid JSON purchase payload.
func validPurchaseBody() map[string]interface{} {
	return map[string]interface{}{
		"name":     "Test Purchase",
		"buyFrom":  "Woolworths",
		"buyAt":    "2026-02-25",
		"buyPrice": "50.00",
		"items":    []interface{}{},
	}
}

// --- HandleCreatePurchase ---

// TestHandleCreatePurchaseSuccess verifies that a valid JSON body results in 201.
func TestHandleCreatePurchaseSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, dbMock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	dbMock.ExpectBegin()
	dbMock.ExpectCommit()
	tx, _ := db.Begin()

	mockRepo := new(MockDocumentRepository)
	mockRepo.On("BeginTx").Return(tx, nil)
	mockRepo.On("SetLatestObjId", mock.Anything, mock.Anything).Return(int64(1), nil)
	mockRepo.On("SaveMetadataWithObjId", mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	mockRepo.On("SaveItems", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	logger, _ := test.NewNullLogger()
	svc := testutil.NewDocumentService(t, mockRepo)
	h := &documentApi.DocumentHandler{DocSvc: *svc, Logger: logger}

	c, w := buildJSONContext("POST", pathPurchases, validPurchaseBody(), 1, nil)
	h.HandleCreatePurchase(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockRepo.AssertExpectations(t)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

// TestHandleCreatePurchaseUnauthorized verifies that missing userId returns 401.
func TestHandleCreatePurchaseUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger, _ := test.NewNullLogger()
	svc := testutil.NewDocumentService(t, new(MockDocumentRepository))
	h := &documentApi.DocumentHandler{DocSvc: *svc, Logger: logger}

	c, w := buildJSONContext("POST", pathPurchases, validPurchaseBody(), 0, nil) // 0 = no userId
	h.HandleCreatePurchase(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestHandleCreatePurchaseInvalidDate verifies that a non-YYYY-MM-DD buyAt returns 400.
func TestHandleCreatePurchaseInvalidDate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body := validPurchaseBody()
	body["buyAt"] = "25/02/2026" // Australian DD/MM/YYYY — wrong format for this endpoint

	logger, _ := test.NewNullLogger()
	svc := testutil.NewDocumentService(t, new(MockDocumentRepository))
	h := &documentApi.DocumentHandler{DocSvc: *svc, Logger: logger}

	c, w := buildJSONContext("POST", pathPurchases, body, 1, nil)
	h.HandleCreatePurchase(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- HandleUpdatePurchase ---

// TestHandleUpdatePurchaseSuccess verifies that a valid PATCH request returns 200.
func TestHandleUpdatePurchaseSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockDocumentRepository)
	dbMock := setupUpdateTx(t, mockRepo, int64(5), 1, int64(10))

	logger, _ := test.NewNullLogger()
	svc := testutil.NewDocumentService(t, mockRepo)
	h := &documentApi.DocumentHandler{DocSvc: *svc, Logger: logger}

	c, w := buildJSONContext("PATCH", pathPurchases5, validPurchaseBody(), 1, gin.Params{{Key: "id", Value: "5"}})
	h.HandleUpdatePurchase(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

// TestHandleUpdatePurchaseNotFound verifies that a missing document returns 404.
func TestHandleUpdatePurchaseNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, dbMock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	dbMock.ExpectBegin()
	dbMock.ExpectRollback()
	tx, _ := db.Begin()

	mockRepo := new(MockDocumentRepository)
	mockRepo.On("BeginTx").Return(tx, nil)
	mockRepo.On("UpdatePurchaseMetadata", mock.Anything, int64(99), 1, mock.Anything).Return(sql.ErrNoRows)

	logger, _ := test.NewNullLogger()
	svc := testutil.NewDocumentService(t, mockRepo)
	h := &documentApi.DocumentHandler{DocSvc: *svc, Logger: logger}

	c, w := buildJSONContext("PATCH", pathPurchases99, validPurchaseBody(), 1, gin.Params{{Key: "id", Value: "99"}})
	h.HandleUpdatePurchase(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockRepo.AssertExpectations(t)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

// TestHandleUpdatePurchaseInvalidID verifies that a non-numeric id returns 400.
func TestHandleUpdatePurchaseInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger, _ := test.NewNullLogger()
	svc := testutil.NewDocumentService(t, new(MockDocumentRepository))
	h := &documentApi.DocumentHandler{DocSvc: *svc, Logger: logger}

	c, w := buildJSONContext("PATCH", pathPurchasesAbc, validPurchaseBody(), 1, gin.Params{{Key: "id", Value: "abc"}})
	h.HandleUpdatePurchase(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestHandleUpdatePurchaseUnauthorized verifies that missing userId returns 401.
func TestHandleUpdatePurchaseUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger, _ := test.NewNullLogger()
	svc := testutil.NewDocumentService(t, new(MockDocumentRepository))
	h := &documentApi.DocumentHandler{DocSvc: *svc, Logger: logger}

	c, w := buildJSONContext("PATCH", pathPurchases5, validPurchaseBody(), 0, gin.Params{{Key: "id", Value: "5"}})
	h.HandleUpdatePurchase(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// setupUpdateWithFileTx extends setupUpdateTx with a UpsertFilePath expectation.
func setupUpdateWithFileTx(t *testing.T, repo *MockDocumentRepository, objId int64, userID int, docId int64) sqlmock.Sqlmock {
	t.Helper()
	dbMock := setupUpdateTx(t, repo, objId, userID, docId)
	repo.On("UpsertFilePath", mock.Anything, mock.Anything).Return(nil)
	return dbMock
}

// buildFormWithFile creates a multipart form body that includes a file field.
func buildFormWithFile(t *testing.T, fields map[string]string, fileContent []byte, fileName string) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	for k, v := range fields {
		_ = w.WriteField(k, v)
	}
	fw, err := w.CreateFormFile("file", fileName)
	assert.NoError(t, err)
	_, _ = fw.Write(fileContent)
	_ = w.Close()
	return body, w.FormDataContentType()
}

// validMultipartFields returns form fields equivalent to validPurchaseBody but
// using DD/MM/YYYY date format (as expected by the multipart path).
func validMultipartFields() map[string]string {
	return map[string]string{
		"name":     "Test Purchase",
		"buyFrom":  "Woolworths",
		"buyAt":    "25/02/2026",
		"buyPrice": "50.00",
	}
}

// --- HandleUpdatePurchase (multipart / with file) ---

// TestHandleUpdatePurchaseWithFileSuccess verifies that a valid multipart PATCH returns 200.
func TestHandleUpdatePurchaseWithFileSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Cleanup(func() { os.RemoveAll("./filedata") })

	mockRepo := new(MockDocumentRepository)
	dbMock := setupUpdateWithFileTx(t, mockRepo, int64(5), 1, int64(10))

	logger, _ := test.NewNullLogger()
	svc := testutil.NewDocumentService(t, mockRepo)
	h := &documentApi.DocumentHandler{DocSvc: *svc, Logger: logger}

	body, contentType := buildFormWithFile(t, validMultipartFields(), []byte("fake pdf"), "receipt.pdf")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("PATCH", pathPurchases5, body)
	c.Request.Header.Set(headerContentType, contentType)
	c.Set("userId", 1)
	c.Params = gin.Params{{Key: "id", Value: "5"}}

	h.HandleUpdatePurchase(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

// TestHandleUpdatePurchaseWithFileInvalidDate verifies that DD/MM/YYYY is required; YYYY-MM-DD returns 400.
func TestHandleUpdatePurchaseWithFileInvalidDate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger, _ := test.NewNullLogger()
	svc := testutil.NewDocumentService(t, new(MockDocumentRepository))
	h := &documentApi.DocumentHandler{DocSvc: *svc, Logger: logger}

	fields := map[string]string{"name": "Receipt", "buyAt": "2026-02-25"} // YYYY-MM-DD — wrong for multipart
	body, contentType := buildFormWithFile(t, fields, []byte("pdf"), "receipt.pdf")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("PATCH", pathPurchases5, body)
	c.Request.Header.Set(headerContentType, contentType)
	c.Set("userId", 1)
	c.Params = gin.Params{{Key: "id", Value: "5"}}

	h.HandleUpdatePurchase(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestHandleUpdatePurchaseWithFileInvalidItems verifies that malformed items JSON returns 400.
func TestHandleUpdatePurchaseWithFileInvalidItems(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger, _ := test.NewNullLogger()
	svc := testutil.NewDocumentService(t, new(MockDocumentRepository))
	h := &documentApi.DocumentHandler{DocSvc: *svc, Logger: logger}

	fields := map[string]string{"name": "Receipt", "items": "not-valid-json"}
	body, contentType := buildFormWithFile(t, fields, []byte("pdf"), "receipt.pdf")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("PATCH", pathPurchases5, body)
	c.Request.Header.Set(headerContentType, contentType)
	c.Set("userId", 1)
	c.Params = gin.Params{{Key: "id", Value: "5"}}

	h.HandleUpdatePurchase(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestHandleUpdatePurchaseWithFileNotFound verifies that a missing document returns 404.
func TestHandleUpdatePurchaseWithFileNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, dbMock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	dbMock.ExpectBegin()
	dbMock.ExpectRollback()
	tx, _ := db.Begin()

	mockRepo := new(MockDocumentRepository)
	mockRepo.On("BeginTx").Return(tx, nil)
	mockRepo.On("UpdatePurchaseMetadata", mock.Anything, int64(99), 1, mock.Anything).Return(sql.ErrNoRows)

	logger, _ := test.NewNullLogger()
	svc := testutil.NewDocumentService(t, mockRepo)
	h := &documentApi.DocumentHandler{DocSvc: *svc, Logger: logger}

	body, contentType := buildFormWithFile(t, map[string]string{"name": "Ghost"}, []byte("pdf"), "receipt.pdf")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("PATCH", pathPurchases99, body)
	c.Request.Header.Set(headerContentType, contentType)
	c.Set("userId", 1)
	c.Params = gin.Params{{Key: "id", Value: "99"}}

	h.HandleUpdatePurchase(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockRepo.AssertExpectations(t)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

// --- HandleGetPurchaseItems ---

// newGetItemsContext builds a test gin.Context for GET /purchases/:id/items.
func newGetItemsContext(userID int, idParam string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/purchases/"+idParam+"/items", nil)
	if userID != 0 {
		c.Set("userId", userID)
	}
	c.Params = gin.Params{{Key: "id", Value: idParam}}
	return c, w
}

// TestHandleGetPurchaseItemsSuccess verifies that items are returned as JSON with 200.
func TestHandleGetPurchaseItemsSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	items := []document.Item{
		{Name: "Coffee beans", Quantity: "2", UnitPrice: "15.00", SubTotal: "30.00"},
		{Name: "Milk", Quantity: "1", UnitPrice: "3.50", SubTotal: "3.50"},
	}

	mockRepo := new(MockDocumentRepository)
	mockRepo.On("GetItemsByPurchaseId", int64(5), 1).Return(items, nil)

	logger, _ := test.NewNullLogger()
	svc := testutil.NewDocumentService(t, mockRepo)
	h := &documentApi.DocumentHandler{DocSvc: *svc, Logger: logger}

	c, w := newGetItemsContext(1, "5")
	h.HandleGetPurchaseItems(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp []documentApi.ItemResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Len(t, resp, 2)
	assert.Equal(t, "Coffee beans", resp[0].ItemName)
	assert.Equal(t, "2", resp[0].ItemQty)
	assert.Equal(t, "15.00", resp[0].UnitPrice)
	assert.Equal(t, "30.00", resp[0].SubTotal)
	assert.Equal(t, "Milk", resp[1].ItemName)
	mockRepo.AssertExpectations(t)
}

// TestHandleGetPurchaseItemsEmpty verifies that an empty slice (no items) returns 200 with [].
func TestHandleGetPurchaseItemsEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockDocumentRepository)
	mockRepo.On("GetItemsByPurchaseId", int64(5), 1).Return([]document.Item{}, nil)

	logger, _ := test.NewNullLogger()
	svc := testutil.NewDocumentService(t, mockRepo)
	h := &documentApi.DocumentHandler{DocSvc: *svc, Logger: logger}

	c, w := newGetItemsContext(1, "5")
	h.HandleGetPurchaseItems(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp []documentApi.ItemResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Empty(t, resp)
	mockRepo.AssertExpectations(t)
}

// TestHandleGetPurchaseItemsUnauthorized verifies that missing userId returns 401.
func TestHandleGetPurchaseItemsUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger, _ := test.NewNullLogger()
	svc := testutil.NewDocumentService(t, new(MockDocumentRepository))
	h := &documentApi.DocumentHandler{DocSvc: *svc, Logger: logger}

	c, w := newGetItemsContext(0, "5") // 0 = no userId set
	h.HandleGetPurchaseItems(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestHandleGetPurchaseItemsInvalidID verifies that a non-numeric id returns 400.
func TestHandleGetPurchaseItemsInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger, _ := test.NewNullLogger()
	svc := testutil.NewDocumentService(t, new(MockDocumentRepository))
	h := &documentApi.DocumentHandler{DocSvc: *svc, Logger: logger}

	c, w := newGetItemsContext(1, "abc")
	h.HandleGetPurchaseItems(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestHandleGetPurchaseItemsNotFound verifies that sql.ErrNoRows returns 404.
func TestHandleGetPurchaseItemsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockDocumentRepository)
	mockRepo.On("GetItemsByPurchaseId", int64(99), 1).Return([]document.Item{}, sql.ErrNoRows)

	logger, _ := test.NewNullLogger()
	svc := testutil.NewDocumentService(t, mockRepo)
	h := &documentApi.DocumentHandler{DocSvc: *svc, Logger: logger}

	c, w := newGetItemsContext(1, "99")
	h.HandleGetPurchaseItems(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockRepo.AssertExpectations(t)
}

// TestHandleGetPurchaseItemsRepoError verifies that a generic repository error returns 500.
func TestHandleGetPurchaseItemsRepoError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockDocumentRepository)
	mockRepo.On("GetItemsByPurchaseId", int64(5), 1).Return([]document.Item{}, errors.New("db failure"))

	logger, _ := test.NewNullLogger()
	svc := testutil.NewDocumentService(t, mockRepo)
	h := &documentApi.DocumentHandler{DocSvc: *svc, Logger: logger}

	c, w := newGetItemsContext(1, "5")
	h.HandleGetPurchaseItems(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockRepo.AssertExpectations(t)
}

// --- HandleDeletePurchase ---

// TestHandleDeletePurchaseSuccess verifies that a valid DELETE returns 204.
func TestHandleDeletePurchaseSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockDocumentRepository)
	mockRepo.On("SoftDeletePurchase", int64(5), 1).Return(nil)

	logger, _ := test.NewNullLogger()
	svc := testutil.NewDocumentService(t, mockRepo)
	h := &documentApi.DocumentHandler{DocSvc: *svc, Logger: logger}

	c, w := buildJSONContext("DELETE", pathPurchases5, nil, 1, gin.Params{{Key: "id", Value: "5"}})
	h.HandleDeletePurchase(c)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockRepo.AssertExpectations(t)
}

// TestHandleDeletePurchaseNotFound verifies that a missing document returns 404.
func TestHandleDeletePurchaseNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockDocumentRepository)
	mockRepo.On("SoftDeletePurchase", int64(99), 1).Return(sql.ErrNoRows)

	logger, _ := test.NewNullLogger()
	svc := testutil.NewDocumentService(t, mockRepo)
	h := &documentApi.DocumentHandler{DocSvc: *svc, Logger: logger}

	c, w := buildJSONContext("DELETE", pathPurchases99, nil, 1, gin.Params{{Key: "id", Value: "99"}})
	h.HandleDeletePurchase(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockRepo.AssertExpectations(t)
}

// TestHandleDeletePurchaseInvalidID verifies that a non-numeric id returns 400.
func TestHandleDeletePurchaseInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger, _ := test.NewNullLogger()
	svc := testutil.NewDocumentService(t, new(MockDocumentRepository))
	h := &documentApi.DocumentHandler{DocSvc: *svc, Logger: logger}

	c, w := buildJSONContext("DELETE", pathPurchasesAbc, nil, 1, gin.Params{{Key: "id", Value: "abc"}})
	h.HandleDeletePurchase(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestHandleDeletePurchaseUnauthorized verifies that missing userId returns 401.
func TestHandleDeletePurchaseUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger, _ := test.NewNullLogger()
	svc := testutil.NewDocumentService(t, new(MockDocumentRepository))
	h := &documentApi.DocumentHandler{DocSvc: *svc, Logger: logger}

	c, w := buildJSONContext("DELETE", pathPurchases5, nil, 0, gin.Params{{Key: "id", Value: "5"}})
	h.HandleDeletePurchase(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// --- HandleGetPurchaseById ---

// newGetPurchaseByIdContext builds a test gin.Context for GET /purchases/:id.
func newGetPurchaseByIdContext(userID int, idParam string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/purchases/"+idParam, nil)
	if userID != 0 {
		c.Set("userId", userID)
	}
	c.Params = gin.Params{{Key: "id", Value: idParam}}
	return c, w
}

// TestHandleGetPurchaseByIdSuccess verifies that a single purchase is returned as JSON with 200.
func TestHandleGetPurchaseByIdSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	buyAt := document.Date{Time: time.Date(2024, 3, 5, 0, 0, 0, 0, time.UTC)}
	doc := &document.Document{
		Id:       "5",
		Title:    "Monthly Groceries",
		FileName: "coles_mar2024.pdf",
		FilePath: "./filedata/0/0/5.pdf",
		PurchaseInfo: document.PurchaseInfo{
			BuyAt:    &buyAt,
			BuyFrom:  "Coles",
			BuyPrice: "112.30",
		},
		Items: []document.Item{
			{Name: "Apple", Quantity: "2", UnitPrice: "1.50", SubTotal: "3.00"},
		},
	}

	mockRepo := new(MockDocumentRepository)
	mockRepo.On("GetPurchaseByObjId", int64(5), 1).Return(doc, nil)

	logger, _ := test.NewNullLogger()
	svc := testutil.NewDocumentService(t, mockRepo)
	h := &documentApi.DocumentHandler{DocSvc: *svc, Logger: logger}

	c, w := newGetPurchaseByIdContext(1, "5")
	h.HandleGetPurchaseById(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp documentApi.PurchaseResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "5", resp.ID)
	assert.Equal(t, "Monthly Groceries", resp.Title)
	assert.Equal(t, "2024-03-05", resp.BuyAt)
	assert.Equal(t, "Coles", resp.BuyFrom)
	assert.Equal(t, "112.30", resp.BuyPrice)
	assert.Equal(t, "/v1/auth/file/5", resp.FileURL)
	assert.Len(t, resp.Items, 1)
	assert.Equal(t, "Apple", resp.Items[0].ItemName)
	mockRepo.AssertExpectations(t)
}

// TestHandleGetPurchaseByIdNoFileURL verifies that fileUrl is omitted when no file is attached.
func TestHandleGetPurchaseByIdNoFileURL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	doc := &document.Document{
		Id:           "5",
		Title:        "No File",
		FilePath:     "",
		PurchaseInfo: document.PurchaseInfo{},
		Items:        []document.Item{},
	}

	mockRepo := new(MockDocumentRepository)
	mockRepo.On("GetPurchaseByObjId", int64(5), 1).Return(doc, nil)

	logger, _ := test.NewNullLogger()
	svc := testutil.NewDocumentService(t, mockRepo)
	h := &documentApi.DocumentHandler{DocSvc: *svc, Logger: logger}

	c, w := newGetPurchaseByIdContext(1, "5")
	h.HandleGetPurchaseById(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp documentApi.PurchaseResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Empty(t, resp.FileURL)
	mockRepo.AssertExpectations(t)
}

// TestHandleGetPurchaseByIdNotFound verifies that sql.ErrNoRows returns 404.
func TestHandleGetPurchaseByIdNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockDocumentRepository)
	mockRepo.On("GetPurchaseByObjId", int64(99), 1).Return(nil, sql.ErrNoRows)

	logger, _ := test.NewNullLogger()
	svc := testutil.NewDocumentService(t, mockRepo)
	h := &documentApi.DocumentHandler{DocSvc: *svc, Logger: logger}

	c, w := newGetPurchaseByIdContext(1, "99")
	h.HandleGetPurchaseById(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockRepo.AssertExpectations(t)
}

// TestHandleGetPurchaseByIdRepoError verifies that a generic repository error returns 500.
func TestHandleGetPurchaseByIdRepoError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockDocumentRepository)
	mockRepo.On("GetPurchaseByObjId", int64(5), 1).Return(nil, errors.New("db failure"))

	logger, _ := test.NewNullLogger()
	svc := testutil.NewDocumentService(t, mockRepo)
	h := &documentApi.DocumentHandler{DocSvc: *svc, Logger: logger}

	c, w := newGetPurchaseByIdContext(1, "5")
	h.HandleGetPurchaseById(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockRepo.AssertExpectations(t)
}

// TestHandleGetPurchaseByIdInvalidID verifies that a non-numeric id returns 400.
func TestHandleGetPurchaseByIdInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger, _ := test.NewNullLogger()
	svc := testutil.NewDocumentService(t, new(MockDocumentRepository))
	h := &documentApi.DocumentHandler{DocSvc: *svc, Logger: logger}

	c, w := newGetPurchaseByIdContext(1, "abc")
	h.HandleGetPurchaseById(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestHandleGetPurchaseByIdUnauthorized verifies that missing userId returns 401.
func TestHandleGetPurchaseByIdUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger, _ := test.NewNullLogger()
	svc := testutil.NewDocumentService(t, new(MockDocumentRepository))
	h := &documentApi.DocumentHandler{DocSvc: *svc, Logger: logger}

	c, w := newGetPurchaseByIdContext(0, "5") // 0 = no userId set
	h.HandleGetPurchaseById(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
