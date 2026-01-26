package authApi

import (
	v1 "2_Go/api/v1"
	"2_Go/internal/auth"
	"2_Go/utils"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type AuthHandler struct {
	AccountSvc auth.AuthService
	DB         *sql.DB // replace with your DB type
	// Logger     *logrus.Logger
	Logger logrus.FieldLogger
}

func NewHandler(acct auth.AuthService, db *sql.DB, logger logrus.FieldLogger) *AuthHandler {
	return &AuthHandler{
		AccountSvc: acct,
		DB:         db,
		Logger:     logger,
	}
}

func (h *AuthHandler) GetPing(c *gin.Context) {
	h.Logger.Debugf("Ping from %s", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

// reset password endpoint.
// authGroup.POST("/users/:username/changepassword", func(c *gin.Context) {

// POST users/changepassword.
// TODO: Checking if new password is the same with old password issue.
func (h *AuthHandler) HandleChangePassword(c *gin.Context) {

	shouldReturn := utils.IsValidUsername(c)
	if shouldReturn {
		return
	}

	// err := v1.ChangePassword(c, h.AccountSvc, h.DB)
	err := v1.ChangePasswordById(c, h.AccountSvc, h.DB)

	if err != nil {
		h.Logger.Errorf("%s Error Change: %s", c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.Logger.Debugf("%s password updated success", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Success"})
}

// validate access token.
// POST /users/me.
// TODO: resolve the return whoami issue with username and userId.
func (h *AuthHandler) GetMe(c *gin.Context) {

	var req struct {
		TokenId string `json:"tokenId"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err})
		return
	}

	jwtUsername, err := v1.IsValidToken(c, h.AccountSvc)

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
func (h *AuthHandler) HandleLogout(c *gin.Context) {

	// authGroup.POST("/logout", func(c *gin.Context) {
	err := v1.Logout(c, h.AccountSvc)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logout"})
}

// POST /refresh.
func (h *AuthHandler) HandleRefresh(c *gin.Context) {
	// r.POST("/v1/auth/refresh", func(c *gin.Context) {
	err := v1.Refresh(c)

	if err != nil {
		h.Logger.Errorf("%s Error Change: %s", c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.Logger.Debugf("%s access updated success", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Success"})
}
