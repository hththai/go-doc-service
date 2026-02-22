package ocrApi

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hththai/ocr"
	"github.com/sirupsen/logrus"
)

// Extractor is the interface satisfied by *ocr.Service.
// It exists so the handler can be tested with a mock.
type Extractor interface {
	ExtractInvoice(ctx context.Context, fileBytes []byte, filename string) (*ocr.Result, error)
}

// Handler handles OCR-related HTTP requests.
type Handler struct {
	OcrSvc Extractor
	Logger logrus.FieldLogger
}

// NewHandler creates a new OCR handler.
func NewHandler(svc Extractor, logger logrus.FieldLogger) *Handler {
	return &Handler{OcrSvc: svc, Logger: logger}
}

// HandleOCR accepts a multipart file, runs OCR on it, and returns structured JSON.
//
// POST /v1/auth/ocr
// Form field: file (required) — PDF or image
func (h *Handler) HandleOCR(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	src, err := fileHeader.Open()
	if err != nil {
		h.Logger.Errorf("OCR: open upload: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	defer src.Close()

	fileBytes, err := io.ReadAll(src)
	if err != nil {
		h.Logger.Errorf("OCR: read upload: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	result, err := h.OcrSvc.ExtractInvoice(c.Request.Context(), fileBytes, fileHeader.Filename)
	if err != nil {
		h.Logger.Errorf("OCR: extract invoice (%s): %v", fileHeader.Filename, err)
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": fmt.Sprintf("OCR processing failed: %v", err),
		})
		return
	}

	h.Logger.Debugf("OCR: processed %s — model: %s, tokens: %d in / %d out",
		fileHeader.Filename, result.Model, result.InputTokens, result.OutputTokens)
	c.JSON(http.StatusOK, result)
}
