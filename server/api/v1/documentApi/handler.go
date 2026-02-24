package documentApi

import (
	"2_Go/internal/document"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type DocumentHandler struct {
	DocSvc document.DocumentService
	Logger logrus.FieldLogger
}

func NewHandler(doc document.DocumentService, logger logrus.FieldLogger) *DocumentHandler {
	return &DocumentHandler{
		DocSvc: doc,
		Logger: logger,
	}
}

// POST /upload.
func (h *DocumentHandler) HandleUpload(c *gin.Context) {

	// Get user ID from context (set by JWTAuthByCookies middleware).
	userIdValue, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "user not authenticated"})
		return
	}
	userId := userIdValue.(int)

	// Parse buyAt from DD/MM/YYYY (Australian date format). Optional — nil if not provided.
	var buyAt *document.Date
	if raw := c.PostForm("buyAt"); raw != "" {
		parsed, err := time.Parse("02/01/2006", raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid buyAt, expected DD/MM/YYYY"})
			return
		}
		d := document.Date{Time: parsed}
		buyAt = &d
	}

	// Parse items from JSON field.
	var items []document.Item
	if rawItems := c.PostForm("items"); rawItems != "" {
		if err := json.Unmarshal([]byte(rawItems), &items); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid items format"})
			return
		}
	}

	// Build upload input.
	input := &document.UploadInput{
		Title:       c.PostForm("name"),
		Description: c.PostForm("description"),
		BuyFrom:     c.PostForm("buyFrom"),
		BuyAt:       buyAt,
		BuyPrice:    c.PostForm("buyPrice"),
		Items:       items,
		UserId:      userId,
	}

	// Get file if attached.
	file, _ := c.FormFile("file")
	input.File = file

	// Use SaveUploadedFile from gin.Context as the file save function.
	saveFunc := func(file *multipart.FileHeader, dst string) error {
		return c.SaveUploadedFile(file, dst)
	}
	err := h.DocSvc.UploadDocument(input, saveFunc)

	if err != nil {
		h.Logger.Errorf("%s Error Upload: %s", c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.Logger.Debugf("%s Upload Success", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Success"})
}
