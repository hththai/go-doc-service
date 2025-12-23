package autho

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

type DefaultObj struct {
	CreatedAt  time.Time `json:"createdAt"`
	ModifiedAt time.Time `json:"modifiedAt"`
}

type ErrorResult struct {
	IsSuccess bool
	Error     error
}

type User struct {
	Username    string     `json:"username" validate:"required,min=1"`
	Email       string     `json:"email" validate:"required,email"`
	Password    string     `json:"password"`
	DefaultObj  DefaultObj `json:"objInfo"`
	ErrorResult ErrorResult
}

func ResultSuccess() ErrorResult {
	return ErrorResult{IsSuccess: true, Error: nil}
}

func ResultError(err error) ErrorResult {
	return ErrorResult{IsSuccess: false, Error: err}
}

var validate = validator.New()

func (u *User) Validate() error {

	if err := validate.Struct(u); err != nil {
		return fmt.Errorf("Invalid user %v", err)
	}

	return nil
}
