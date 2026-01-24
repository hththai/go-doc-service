package authoRoute

import (
	v1 "2_Go/api/v1"
	"2_Go/internal/auth"
	"2_Go/internal/document"
	"2_Go/middleware/authen"
	"2_Go/utils"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	AccountSvc auth.AuthService
	DocSvc     document.DocumentService
	DB         *sql.DB // replace with your DB type
}

func NewHandler(acct auth.AuthService, doc document.DocumentService, db *sql.DB) *Handler {
	return &Handler{
		AccountSvc: acct,
		DocSvc:     doc,
		DB:         db,
	}
}

func (h *Handler) getPing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

// reset password endpoint.
// authGroup.POST("/users/:username/changepassword", func(c *gin.Context) {

// POST users/:username/changePassword.
// TODO: record log.
func (h *Handler) handleChangePassword(c *gin.Context) {

	shouldReturn := utils.IsValidUser(c)
	if shouldReturn {
		return
	}

	err := v1.ChangePassword(c, &h.AccountSvc, h.DB)

	if err != nil {
		// log.Errorf("%s Error Change: %s", c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// log.Debugf("%s password updated success", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Success"})
}

// validate access token.
// POST /users/me.
func (h *Handler) getMe(c *gin.Context) {

	var req struct {
		TokenId string `json:"tokenId"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err})
		return
	}

	jwtUsername, err := v1.IsValidToken(c, &h.AccountSvc)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err})
		return
	}
	// check if the token match the current user session in local host.
	if req.TokenId != jwtUsername {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid access"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"sessionId": jwtUsername})
}

// POST /logout.
func (h *Handler) handleLogout(c *gin.Context) {

	// authGroup.POST("/logout", func(c *gin.Context) {
	err := v1.Logout(c, &h.AccountSvc)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logout"})
}

// POST /upload.
// TODO: record log.
func (h *Handler) handleUpload(c *gin.Context) {

	tokenId := c.PostForm("tokenId")

	if tokenId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "missing tokenId"})
		return
	}
	jwtUsername, err := v1.IsValidToken(c, &h.AccountSvc)
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
		// log.Errorf("%s Error Upload: %s", c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// log.Debugf("%s Upload Success", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Success"})

}

// authGroup := r.Group("/v1/auth", authen.JWTAuthByCookies())
func RegisterRoutes(r *gin.Engine, h *Handler) {
	authGroup := r.Group("/v1/auth", authen.JWTAuthByCookies())

	authGroup.GET("/test", h.getPing)
	authGroup.GET("/users/me", h.getMe)
	authGroup.POST("/logout", h.handleLogout)
	authGroup.POST("/upload", h.handleUpload)
	authGroup.POST("users/:username/changePassword", h.handleChangePassword)
}
