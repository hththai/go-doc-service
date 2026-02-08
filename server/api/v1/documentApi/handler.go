package documentApi

import (
	"2_Go/internal/auth"
	"2_Go/internal/document"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type DocumentHandler struct {
	AccountSvc auth.AuthService
	DocSvc     document.DocumentService
	Logger     logrus.FieldLogger
}

func NewHandler(acct auth.AuthService, doc document.DocumentService, logger logrus.FieldLogger) *DocumentHandler {
	return &DocumentHandler{
		AccountSvc: acct,
		DocSvc:     doc,
		Logger:     logger,
	}
}

// POST /upload.
func (h *DocumentHandler) HandleUpload(c *gin.Context) {

	tokenId := c.PostForm("tokenId")

	if tokenId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "missing tokenId"})
		return
	}

	// Get access token from cookie
	accessToken, err := c.Cookie("access_token")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "missing access token"})
		return
	}

	// Validate token
	jwtTokenId, err := h.AccountSvc.ValidateAccessToken(accessToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// check if the token match the current user session in local host.
	if tokenId != jwtTokenId {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid access"})
		return
	}

	// Get user ID from context (set by middleware).
	userIdValue, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "user not authenticated"})
		return
	}
	userId := userIdValue.(int)

	// Build upload input.
	input := &document.UploadInput{
		Title:       c.PostForm("name"),
		Description: c.PostForm("description"),
		UserId:      userId,
	}

	// Get file if attached.
	file, _ := c.FormFile("file")
	input.File = file

	// Use SaveUploadedFile from gin.Context as the file save function.
	saveFunc := func(file *multipart.FileHeader, dst string) error {
		return c.SaveUploadedFile(file, dst)
	}
	err = h.DocSvc.UploadDocument(input, saveFunc)

	if err != nil {
		h.Logger.Errorf("%s Error Upload: %s", c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.Logger.Debugf("%s Upload Success", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Success"})
}
