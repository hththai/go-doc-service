package ocrApi_test

import (
	"2_Go/api/v1/ocrApi"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hththai/ocr"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockExtractor satisfies ocrApi.Extractor using testify/mock.
type mockExtractor struct {
	mock.Mock
}

func (m *mockExtractor) ExtractInvoice(ctx context.Context, fileBytes []byte, filename string) (*ocr.Result, error) {
	args := m.Called(ctx, fileBytes, filename)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ocr.Result), args.Error(1)
}

// buildFileForm builds a multipart/form-data body containing a single file field.
func buildFileForm(fieldName, filename string, content []byte) (*bytes.Buffer, string) {
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	part, _ := w.CreateFormFile(fieldName, filename)
	part.Write(content)
	w.Close()
	return body, w.FormDataContentType()
}

func TestHandleOCR(t *testing.T) {
	gin.SetMode(gin.TestMode)

	successResult := &ocr.Result{
		Invoice: ocr.Invoice{
			Seller: "Acme Pty Ltd",
			ABN:    "12 345 678 901",
			Total:  "$1,234.00",
			Items: []ocr.Item{
				{Description: "Widget", Qty: 2, UnitPrice: "$500.00", GSTRate: "10%", Subtotal: "$1,000.00"},
			},
		},
		Model:        "claude-haiku-4-5",
		InputTokens:  800,
		OutputTokens: 120,
		RawResponse:  `{"seller":"Acme Pty Ltd"}`,
	}

	fakeImageBytes := []byte("fake-png-content")

	tests := []struct {
		name           string
		formField      string // form field name ("file" or something wrong)
		filename       string
		fileContent    []byte
		mockSetup      func(m *mockExtractor)
		expectedStatus int
		checkBody      func(t *testing.T, body []byte)
	}{
		{
			name:           "no file field returns 400",
			formField:      "wrong_field",
			filename:       "invoice.png",
			fileContent:    fakeImageBytes,
			mockSetup:      func(m *mockExtractor) {},
			expectedStatus: http.StatusBadRequest,
			checkBody: func(t *testing.T, body []byte) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(body, &resp))
				assert.Equal(t, "file is required", resp["error"])
			},
		},
		{
			name:        "successful OCR returns 200 with invoice only",
			formField:   "file",
			filename:    "invoice.png",
			fileContent: fakeImageBytes,
			mockSetup: func(m *mockExtractor) {
				m.On("ExtractInvoice", mock.Anything, fakeImageBytes, "invoice.png").
					Return(successResult, nil)
			},
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var resp ocrApi.OCRResponse
				assert.NoError(t, json.Unmarshal(body, &resp))
				assert.Equal(t, "Acme Pty Ltd", resp.Invoice.Seller)
				assert.Equal(t, "12 345 678 901", resp.Invoice.ABN)
				assert.Equal(t, "$1,234.00", resp.Invoice.Total)
				assert.Len(t, resp.Invoice.Items, 1)

				// Internal fields must not be exposed to callers.
				var raw map[string]any
				assert.NoError(t, json.Unmarshal(body, &raw))
				assert.NotContains(t, raw, "model")
				assert.NotContains(t, raw, "input_tokens")
				assert.NotContains(t, raw, "output_tokens")
				assert.NotContains(t, raw, "raw_response")
			},
		},
		{
			name:        "OCR service error returns 422 with error message",
			formField:   "file",
			filename:    "invoice.pdf",
			fileContent: fakeImageBytes,
			mockSetup: func(m *mockExtractor) {
				m.On("ExtractInvoice", mock.Anything, fakeImageBytes, "invoice.pdf").
					Return(nil, errors.New("calling Claude API: network timeout"))
			},
			expectedStatus: http.StatusUnprocessableEntity,
			checkBody: func(t *testing.T, body []byte) {
				var resp map[string]string
				assert.NoError(t, json.Unmarshal(body, &resp))
				assert.Contains(t, resp["error"], "OCR processing failed")
				assert.Contains(t, resp["error"], "network timeout")
			},
		},
		{
			name:        "PDF file is forwarded correctly",
			formField:   "file",
			filename:    "receipt.pdf",
			fileContent: []byte("%PDF-1.4 fake pdf bytes"),
			mockSetup: func(m *mockExtractor) {
				m.On("ExtractInvoice", mock.Anything, []byte("%PDF-1.4 fake pdf bytes"), "receipt.pdf").
					Return(successResult, nil)
			},
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var resp ocrApi.OCRResponse
				assert.NoError(t, json.Unmarshal(body, &resp))
				assert.Equal(t, "Acme Pty Ltd", resp.Invoice.Seller)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(mockExtractor)
			tt.mockSetup(svc)

			logger, _ := test.NewNullLogger()
			h := &ocrApi.Handler{OcrSvc: svc, Logger: logger}

			body, contentType := buildFileForm(tt.formField, tt.filename, tt.fileContent)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodPost, "/v1/auth/ocr", body)
			c.Request.Header.Set("Content-Type", contentType)

			h.HandleOCR(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkBody(t, w.Body.Bytes())
			svc.AssertExpectations(t)
		})
	}
}
