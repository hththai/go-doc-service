package auth

import (
	"2_Go/internal/obj"
	"fmt"

	"github.com/go-playground/validator/v10"
)

type Account struct {
	UserId      int            `json:"userId"`
	Username    string         `json:"username" binding:"required" validate:"required"`
	Email       string         `json:"email"`
	Password    string         `json:"password" binding:"required,min=5" validate:"required,min=5"`
	DefaultObj  obj.DefaultObj `json:"objDefault"`
	ErrorResult obj.ErrorResult
}

var validate = validator.New()

// Simple validate
func (r *Account) Validate() error {
	if err := validate.Struct(r); err != nil {
		return fmt.Errorf("Invalid account %v", err)
	}

	return nil
}

// Validate Password only.
func (r *Account) ValidatePassword() error {
	if err := validate.StructPartial(r, "Password"); err != nil {
		return fmt.Errorf("Invalid password %v", err)
	}

	return nil
}
