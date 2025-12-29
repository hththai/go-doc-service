package v1

import (
	"2_Go/internal/auth"
	"2_Go/internal/obj"
	"2_Go/middleware/authen"

	"database/sql"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handle register request with username and password.
func Register(c *gin.Context, service *auth.AuthService, db *sql.DB) error {

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

// Change password.
// TODO: check the validated account or not with matching username and password.
// TODO: Do we need password field. or JWT. if yes, handle validate correct account with current username and password.
func ChangePassword(c *gin.Context, service *auth.AuthService, db *sql.DB) error {
	var req struct {
		Username    string `json:"username"`
		Password    string `json:"password"`
		NewPassword string `json:"newpassword"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}

	// bind to account.
	var account auth.Account
	account.Username = req.Username
	account.Password = req.Password

	// Validate current user by password.
	if err := service.ValidateAccountService(db, account.Username, account.Password); err != nil {
		return fmt.Errorf("incorrect login")
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
func Login(c *gin.Context, service *auth.AuthService, db *sql.DB) error {
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

// Login with jwt
func LoginJwt(c *gin.Context, service *auth.AuthService, db *sql.DB) (string, error) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		return "", err
	}

	account, err := service.Login(db, req.Username, req.Password)

	if err != nil {
		return "", err
	}

	// Assign account username.
	account.Username = req.Username

	// issue token.
	token, err := authen.CreateToken(account.Username)

	if err != nil {
		return "", err
	}

	return token, nil
}
