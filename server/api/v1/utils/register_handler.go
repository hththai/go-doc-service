package utils

import (
	"2_Go/internal/auth"
	"2_Go/internal/obj"
	"2_Go/middleware/authen"
	"net/http"

	"database/sql"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handle register request with username and password.
func Register(c *gin.Context, service auth.AuthService, db *sql.DB) error {

	var account auth.Account

	if err := c.ShouldBindJSON(&account); err != nil {
		return err
	}

	account.DefaultObj = obj.DefaultObj{GUID: uuid.New().String()}

	err := account.Validate()
	if err != nil {
		return err
	}

	// Begin transaction.
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	// Always ensure roll back if something goes wrong.
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		}
	}()

	// Call service.
	_, err = service.Register(tx, account)

	// Check if register fail
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	return nil
}

// TODO: deleting
// Change password.
// func ChangePassword(c *gin.Context, service auth.AuthService, db *sql.DB) error {

// 	var req struct {
// 		// Username    string `json:"username"`
// 		Password    string `json:"password"`
// 		NewPassword string `json:"newpassword"`
// 	}

// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		return err
// 	}

// 	// Prevent wrong username.
// 	reqUser := c.Param("username")

// 	// bind to account.
// 	var account auth.Account
// 	// account.Username = req.Username
// 	account.Username = reqUser
// 	account.Password = req.Password

// 	// Validate current user by password.
// 	if err := service.ValidateAccountService(db, account.Username, account.Password); err != nil {
// 		return fmt.Errorf("incorrect login")
// 	}

// 	// Begin transaction.
// 	tx, err := db.Begin()
// 	if err != nil {
// 		return fmt.Errorf("failed to start transaction: %w", err)
// 	}

// 	// Always ensure roll back if something goes wrong.
// 	defer func() {
// 		if p := recover(); p != nil {
// 			tx.Rollback()
// 			panic(p)
// 		} else if err != nil {
// 			tx.Rollback()
// 		}
// 	}()

// 	// Call update pasword service.
// 	_, err = service.ChangePasswordService(tx, account, req.NewPassword)
// 	// Check if register fail
// 	if err != nil {
// 		return err
// 	}

// 	if err = tx.Commit(); err != nil {
// 		return fmt.Errorf("commit failed: %w", err)
// 	}

// 	return nil
// }

func ChangePasswordById(c *gin.Context, service auth.AuthService, db *sql.DB) error {

	var req struct {
		// Username    string `json:"username"`
		Password    string `json:"password"`
		NewPassword string `json:"newpassword"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}

	// Prevent wrong account
	reqUser, exists := c.Get("userId")

	if !exists {
		return fmt.Errorf("Invalid user id")
	}

	// bind to account.
	var account auth.Account
	// account.Username = req.Username
	account.UserId = reqUser.(int)
	account.Password = req.Password

	// Validate current user by password.
	// if err := service.ValidateAccountService(db, account.Username, account.Password); err != nil {
	// 	return fmt.Errorf("incorrect login")
	// }

	if err := service.ValidateAccountByIdService(db, account.UserId, account.Password); err != nil {
		// return fmt.Errorf("incorrect login")
		return err
	}

	// Begin transaction.
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	// Always ensure roll back if something goes wrong.
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		}
	}()

	// Call update pasword service.
	_, err = service.ChangePasswordService(tx, account, req.NewPassword)
	// Check if register fail
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	return nil
}

// Login
func Login(c *gin.Context, service auth.AuthService, db *sql.DB) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		return err
	}

	_, err := service.Login(db, req.Username, req.Password)

	if err != nil {
		return err
	}

	return nil
}

// Test return with token object
func LoginJwt(c *gin.Context, service auth.AuthService, db *sql.DB) (obj.AuthToken, error) {
	var tokenReturn obj.AuthToken

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		return tokenReturn, err
	}

	account, err := service.Login(db, req.Username, req.Password)

	if err != nil {
		return tokenReturn, err
	}

	// Update userId to return token value.
	tokenReturn.UserId = account.UserId

	// Assign account username.
	account.Username = req.Username

	// issue accessToken.
	_, err = handleAccessToken(account.UserId, &tokenReturn, c)
	if err != nil {
		return tokenReturn, err
	}

	// Assign token value.

	// Handle refresh token
	_ = handleRefreshToken(account.UserId, c)

	// c.SetCookie("jwt", token, 3600, "/", "localhost", true, true)

	return tokenReturn, nil
}

// Logout with jwt
func Logout(c *gin.Context, service auth.AuthService) error {
	err := handleLogout(c)
	if err != nil {
		return fmt.Errorf("Error of log out")
	}

	return nil
}

// Clean access_token and refresh_token
func handleLogout(c *gin.Context) error {
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
	return nil
}

// Validate access token.
func IsValidToken(c *gin.Context, service auth.AuthService) (string, error) {
	accessToken, err := c.Cookie("access_token")
	if err != nil {
		return "", fmt.Errorf("missing access token")
	}

	claims, err := authen.VerifyToken(accessToken)

	if err != nil {
		return "", fmt.Errorf("invalid token")
	}

	// Claim token.

	// tokenId, ok := (*claims)["tokenId"].(string)
	tokenId, ok := (*claims)["tokenId"].(string)
	if !ok {
		return "", fmt.Errorf("tokenId not found in token")
	}

	return tokenId, nil
}

// TODO: Fix the refresh token.
func Refresh(c *gin.Context) error {
	// fmt.Println("Cookies:::", c.Request.Cookies())
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		return fmt.Errorf("missing refresh token")
	}

	claims, err := authen.VerifyToken(refreshToken)
	if err != nil {
		return fmt.Errorf("invalid refresh token")
	}

	raw := (*claims)["userId"]
	userIdFloat, ok := raw.(float64)
	if !ok {
		return fmt.Errorf("invalid userID type")
	}

	userId := int(userIdFloat)

	var newTokenReturn obj.AuthToken

	// Issue new token.
	_, err = handleAccessToken(userId, &newTokenReturn, c)

	return nil
}

// Handle refresh token when access token invalid.
// func Refresh(c *gin.Context) error {
// 	// fmt.Println("Cookies:::", c.Request.Cookies())
// 	refreshToken, err := c.Cookie("refresh_token")
// 	if err != nil {
// 		return fmt.Errorf("missing refresh token")
// 	}

// 	claims, err := authen.VerifyToken(refreshToken)
// 	if err != nil {
// 		return fmt.Errorf("invalid refresh token")
// 	}

// 	username := (*claims)["username"].(string)

// 	// Issue new token.
// 	_, err = handleAccessToken(username, c)

// 	return nil
// }

func handleAccessToken(userId int, tokenReturn *obj.AuthToken, c *gin.Context) (*obj.AuthToken, error) {
	token, tokenId, err := authen.CreateAccessToken(userId)
	if err != nil {
		return nil, err
	}

	tokenReturn.AccessToken = token
	tokenReturn.TokenId = tokenId

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "access_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, //false if localhost
		// SameSite: http.SameSiteStrictMode,
		// SameSite: http.SameSiteLaxMode,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600, // 10 min
	})
	return tokenReturn, nil
}

// func handleAccessToken(username string, c *gin.Context) (string, error) {
// 	token, err := authen.CreateAccessToken(username)
// 	if err != nil {
// 		return "", err
// 	}

// 	http.SetCookie(c.Writer, &http.Cookie{
// 		Name:     "access_token",
// 		Value:    token,
// 		Path:     "/",
// 		HttpOnly: true,
// 		Secure:   false,
// 		SameSite: http.SameSiteStrictMode,
// 		MaxAge:   604800,
// 	})
// 	return token, nil
// }

// Set refreshtoken and secure for cookies access.
func handleRefreshToken(userId int, c *gin.Context) error {
	refreshToken, err := authen.CreateRefreshToken(userId)

	if err != nil {
		return fmt.Errorf("login failed with token")
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // for testing purpose without https
		SameSite: http.SameSiteLaxMode,
		// SameSite: http.SameSiteNoneMode,
		MaxAge: 604800, // 1 week
	})

	return nil
}
