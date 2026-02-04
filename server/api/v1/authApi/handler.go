package authApi

import (
	"2_Go/internal/auth"
	"2_Go/internal/obj"
	"2_Go/utils"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

// Cookie helper methods

// setAccessTokenCookie sets the access token as an HTTP-only cookie.
func (h *AuthHandler) setAccessTokenCookie(c *gin.Context, token string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "access_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600, // 10 min
	})
}

// setRefreshTokenCookie sets the refresh token as an HTTP-only cookie.
func (h *AuthHandler) setRefreshTokenCookie(c *gin.Context, token string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   604800, // 1 week
	})
}

// clearAuthCookies clears both access and refresh token cookies.
func (h *AuthHandler) clearAuthCookies(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}

// HandleRegister handles user registration.
// POST /register
func (h *AuthHandler) HandleRegister(c *gin.Context) {
	var account auth.Account
	if err := c.ShouldBindJSON(&account); err != nil {
		h.Logger.Errorf("%s Error binding JSON: %s", c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account.DefaultObj = obj.DefaultObj{GUID: uuid.New().String()}

	err := h.AccountSvc.RegisterWithTransaction(h.DB, account)
	if err != nil {
		h.Logger.Errorf("%s Error Register: %s", c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.Logger.Debugf("%s Register success", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Success"})
}

// HandleLogin handles user login with JWT.
// POST /login
func (h *AuthHandler) HandleLogin(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.Logger.Errorf("%s Error binding JSON: %s", c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Authenticate user
	account, err := h.AccountSvc.Login(h.DB, req.Username, req.Password)
	if err != nil {
		h.Logger.Errorf("%s Error Login: %s", c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create tokens
	accessToken, tokenId, refreshToken, err := h.AccountSvc.CreateTokensForUser(account.UserId)
	if err != nil {
		h.Logger.Errorf("%s Error creating tokens: %s", c.ClientIP(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}

	// Set cookies
	h.setAccessTokenCookie(c, accessToken)
	h.setRefreshTokenCookie(c, refreshToken)

	h.Logger.Debugf("%s Login success", c.ClientIP())
	c.JSON(http.StatusOK, obj.AuthToken{
		UserId:      account.UserId,
		TokenId:     tokenId,
		AccessToken: accessToken,
	})
}

func (h *AuthHandler) GetPing(c *gin.Context) {
	h.Logger.Debugf("Ping from %s", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

// POST /user/changepassword
func (h *AuthHandler) HandleChangePassword(c *gin.Context) {
	shouldReturn := utils.IsValidUsername(c)
	if shouldReturn {
		return
	}

	var req struct {
		Password    string `json:"password"`
		NewPassword string `json:"newpassword"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.Logger.Errorf("%s Error binding JSON: %s", c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get userId from context (set by middleware)
	reqUser, exists := c.Get("userId")
	if !exists {
		h.Logger.Errorf("%s Missing userId in context", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user id"})
		return
	}

	userId := reqUser.(int)

	// Change password with transaction
	err := h.AccountSvc.ChangePasswordWithTransaction(h.DB, userId, req.Password, req.NewPassword)
	if err != nil {
		h.Logger.Errorf("%s Error Change: %s", c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.Logger.Debugf("%s password updated success", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Success"})
}

// GET /users/me - validate access token.
func (h *AuthHandler) GetMe(c *gin.Context) {
	var req struct {
		TokenId string `json:"tokenId"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Get cookie
	accessToken, err := c.Cookie("access_token")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "missing access token"})
		return
	}

	// Validate token
	tokenId, err := h.AccountSvc.ValidateAccessToken(accessToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Check if the token matches
	if req.TokenId != tokenId {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid access"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"sessionId": tokenId})
}

// POST /logout.
func (h *AuthHandler) HandleLogout(c *gin.Context) {
	h.clearAuthCookies(c)
	c.JSON(http.StatusOK, gin.H{"message": "logout"})
}

// POST /refresh.
func (h *AuthHandler) HandleRefresh(c *gin.Context) {
	// Get refresh token cookie
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		h.Logger.Errorf("%s Missing refresh token", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing refresh token"})
		return
	}

	// Refresh the token
	newAccessToken, _, _, err := h.AccountSvc.RefreshAccessToken(refreshToken)
	if err != nil {
		h.Logger.Errorf("%s Error refreshing token: %s", c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set new cookie
	h.setAccessTokenCookie(c, newAccessToken)

	h.Logger.Debugf("%s access token refreshed", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Success"})
}
