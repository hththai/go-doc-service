package main

import (
	"fmt"
	"go_ocr/internal/autho"
)

type Result struct {
	IsSuccess bool
	Error     error
}

type Chain struct {
	stringValue string
	Result      Result
}

// Constructor and return Object.
func NewChain() *Chain {
	return &Chain{
		Result: Result{
			IsSuccess: false,
			Error:     nil,
		},
	}
}

func (c *Result) SuccessResult() Result {
	return Result{
		IsSuccess: true,
		Error:     nil,
	}
}

func (c *Result) ErrorResult(err error) Result {
	return Result{
		IsSuccess: false,
		Error:     err,
	}
}

// Constructor.
func (c *Chain) NewValue(s string) (*Chain, error) {

	if len(s) > 5 {
		c.Result = Result{
			IsSuccess: false,
			Error:     fmt.Errorf("len is too long"),
		}

		return c, c.Result.Error
	}

	c.Result.IsSuccess = true
	return c, nil
}

func (c *Chain) Value() (*Chain, error) {
	if c == nil || c.Result.IsSuccess != true {
		return c, c.Result.Error
	}
	return c, nil
}

func (c *Chain) SetValue(s string) *Chain {

	if len(s) > 5 {
		c.Result = c.Result.ErrorResult(fmt.Errorf("input too long"))
		return c
	}

	c.stringValue = s
	c.Result = c.Result.SuccessResult()

	return c
}

func main() {
	authoRepo := autho.NewAuthoRepo()
	authoService := autho.NewAuthoService(&authoRepo)

	newAccount, err := authoService.RegisterService(autho.User{Email: "hello", Username: "hello", Password: "Hello"})

	if err != nil {
		fmt.Println("Cannot create an account because:", newAccount.ErrorResult.Error)
		return
	}

	fmt.Println("the current status is:::", newAccount.Username)
	fmt.Println("Password is:::", newAccount.Password)

	fmt.Println("Change Password")

	_, err = authoService.ChangePasswordService(&newAccount, "123")
	if err != nil {
		fmt.Println("Cannot change password:::", err)
		return
	}

	fmt.Println("New password is:::", newAccount.Password)

	// inputPass, _ := authoService.HashPassword("Hello")
	// fmt.Println("inputPassword is :::", inputPass)

	_, err = authoService.ValidatePassword(&newAccount, "123")
	if err != nil {
		fmt.Println("Password NOT match")
		return
	}

	fmt.Println("Current password is:", newAccount.Password)

	fmt.Println("Password matched")
}
