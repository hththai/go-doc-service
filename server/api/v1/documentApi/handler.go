package documentApi

import (
	"2_Go/internal/document"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// PurchaseResponse is the JSON shape returned by GET /purchases.
// Fields are intentionally flat to match the client-side Purchase type.
type PurchaseResponse struct {
	ID       string         `json:"id"`
	Title    string         `json:"title"`
	Filename string         `json:"filename"`
	FileURL  string         `json:"fileUrl,omitempty"`
	BuyAt    string         `json:"buyAt"`
	BuyFrom  string         `json:"buyFrom"`
	BuyPrice string         `json:"buyPrice"`
	Items    []ItemResponse `json:"items"`
}

// ItemResponse mirrors the client-side Item type.
type ItemResponse struct {
	ItemName  string `json:"itemName"`
	ItemQty   string `json:"itemQty"`
	UnitPrice string `json:"unitPrice"`
	SubTotal  string `json:"subTotal"`
}

// toPurchaseResponse maps a domain Document to the API response shape.
func toPurchaseResponse(doc document.Document) PurchaseResponse {
	r := PurchaseResponse{
		ID:       doc.Id,
		Title:    doc.Title,
		Filename: doc.FileName,
		BuyFrom:  doc.PurchaseInfo.BuyFrom,
		BuyPrice: doc.PurchaseInfo.BuyPrice,
		Items:    []ItemResponse{},
	}
	if doc.PurchaseInfo.BuyAt != nil {
		r.BuyAt = doc.PurchaseInfo.BuyAt.Format("2006-01-02")
	}
	if doc.FilePath != "" {
		r.FileURL = "/v1/auth/file/" + doc.Id
	}
	for _, item := range doc.Items {
		r.Items = append(r.Items, ItemResponse{
			ItemName:  item.Name,
			ItemQty:   item.Quantity,
			UnitPrice: item.UnitPrice,
			SubTotal:  item.SubTotal,
		})
	}
	return r
}

const (
	msgNotAuthenticated = "user not authenticated"
	msgInvalidID        = "invalid id"
	msgNotFound         = "not found"
)

// purchaseJSONRequest is the JSON body accepted by the create and update endpoints.
type purchaseJSONRequest struct {
	Name     string          `json:"name"`
	BuyFrom  string          `json:"buyFrom"`
	BuyAt    string          `json:"buyAt"` // YYYY-MM-DD from a date input
	BuyPrice string          `json:"buyPrice"`
	Items    []document.Item `json:"items"`
}

// parsePurchaseJSONRequest binds the request body and converts buyAt to a *document.Date.
func parsePurchaseJSONRequest(c *gin.Context) (*purchaseJSONRequest, *document.Date, bool) {
	var req purchaseJSONRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return nil, nil, false
	}
	var buyAt *document.Date
	if req.BuyAt != "" {
		parsed, err := time.Parse("2006-01-02", req.BuyAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid buyAt, expected YYYY-MM-DD"})
			return nil, nil, false
		}
		d := document.Date{Time: parsed}
		buyAt = &d
	}
	return &req, buyAt, true
}

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
		c.JSON(http.StatusUnauthorized, gin.H{"message": msgNotAuthenticated})
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

// GET /purchases?year=2024&month=03
// Returns all active purchases for the authenticated user.
// year and month are optional; pass "all" or omit to return everything.
func (h *DocumentHandler) HandleGetPurchases(c *gin.Context) {
	userIdValue, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": msgNotAuthenticated})
		return
	}
	userId := userIdValue.(int)

	year := c.Query("year")
	month := c.Query("month")

	docs, err := h.DocSvc.GetPurchases(userId, year, month)
	if err != nil {
		h.Logger.Errorf("%s Error GetPurchases: %s", c.ClientIP(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch purchases"})
		return
	}

	resp := make([]PurchaseResponse, 0, len(docs))
	for _, doc := range docs {
		resp = append(resp, toPurchaseResponse(doc))
	}
	c.JSON(http.StatusOK, resp)
}

// GET /file/:id
// Serves the receipt file attached to a purchase. The :id is the document's obj_id.
// Only the owning user may access the file.
func (h *DocumentHandler) HandleServeFile(c *gin.Context) {
	userIdValue, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": msgNotAuthenticated})
		return
	}
	userId := userIdValue.(int)

	objId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": msgInvalidID})
		return
	}

	filePath, fileName, err := h.DocSvc.GetFilePath(objId, userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": msgNotFound})
		} else {
			h.Logger.Errorf("%s Error GetFilePath id=%d: %s", c.ClientIP(), objId, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve file"})
		}
		return
	}
	if filePath == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "no file attached"})
		return
	}

	f, err := os.Open(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
		return
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}

	mimeType := mime.TypeByExtension(filepath.Ext(fileName))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	c.DataFromReader(http.StatusOK, stat.Size(), mimeType, f, map[string]string{
		"Content-Disposition": fmt.Sprintf(`inline; filename="%s"`, fileName),
	})
}

// POST /purchases – create a purchase without a file attachment (JSON body).
func (h *DocumentHandler) HandleCreatePurchase(c *gin.Context) {
	userIdValue, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": msgNotAuthenticated})
		return
	}
	userId := userIdValue.(int)

	req, buyAt, ok := parsePurchaseJSONRequest(c)
	if !ok {
		return
	}

	input := &document.UploadInput{
		Title:    req.Name,
		BuyFrom:  req.BuyFrom,
		BuyAt:    buyAt,
		BuyPrice: req.BuyPrice,
		Items:    req.Items,
		UserId:   userId,
	}

	if err := h.DocSvc.UploadDocument(input, nil); err != nil {
		h.Logger.Errorf("%s Error CreatePurchase: %s", c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Success"})
}

// PATCH /purchases/:id – update metadata and items of an existing purchase (JSON body).
func (h *DocumentHandler) HandleUpdatePurchase(c *gin.Context) {
	userIdValue, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": msgNotAuthenticated})
		return
	}
	userId := userIdValue.(int)

	objId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": msgInvalidID})
		return
	}

	req, buyAt, ok := parsePurchaseJSONRequest(c)
	if !ok {
		return
	}

	input := &document.UploadInput{
		Title:    req.Name,
		BuyFrom:  req.BuyFrom,
		BuyAt:    buyAt,
		BuyPrice: req.BuyPrice,
		Items:    req.Items,
		UserId:   userId,
	}

	if err := h.DocSvc.UpdatePurchase(objId, userId, input); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": msgNotFound})
			return
		}
		h.Logger.Errorf("%s Error UpdatePurchase id=%d: %s", c.ClientIP(), objId, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update purchase"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Success"})
}

// DELETE /purchases/:id – soft-deletes a purchase (sets status=-1).
func (h *DocumentHandler) HandleDeletePurchase(c *gin.Context) {
	userIdValue, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": msgNotAuthenticated})
		return
	}
	userId := userIdValue.(int)

	objId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": msgInvalidID})
		return
	}

	if err := h.DocSvc.DeletePurchase(objId, userId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": msgNotFound})
			return
		}
		h.Logger.Errorf("%s Error DeletePurchase id=%d: %s", c.ClientIP(), objId, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete purchase"})
		return
	}

	c.Status(http.StatusNoContent)
	c.Writer.WriteHeaderNow()
}
