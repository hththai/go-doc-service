package v1

import (
	"2_Go/internal/auth"
	"2_Go/internal/obj"
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
