package documentApi

import (
	v1 "2_Go/api/v1/utils"
	"2_Go/internal/auth"
	"2_Go/internal/document"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type DocumentHandler struct {
	AccountSvc auth.AuthService
	DocSvc     document.DocumentService
	DB         *sql.DB // replace with your DB type
	Logger     logrus.FieldLogger
}

func NewHandler(acct auth.AuthService, doc document.DocumentService, db *sql.DB, logger logrus.FieldLogger) *DocumentHandler {
	return &DocumentHandler{
		AccountSvc: acct,
		DocSvc:     doc,
		DB:         db,
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
	jwtUsername, err := v1.IsValidToken(c, h.AccountSvc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err})
		return
	}
	// check if the token match the current user session in local host.
	if tokenId != jwtUsername {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid access"})
		return
	}
	err = v1.UploadDocument(c, &h.DocSvc, h.DB)

	if err != nil {
		h.Logger.Errorf("%s Error Upload: %s", c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.Logger.Debugf("%s Upload Success", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Success"})
}
