package autho

import (
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type AuthoService struct {
	repo AuthoRepo
}

func NewAuthoService(repo AuthoRepo) *AuthoService {
	return &AuthoService{repo: repo}
}

func (s *AuthoService) RegisterService(user User) (User, error) {
	if !s.isValidUserAccount(&user) {
		return User{ErrorResult: user.ErrorResult}, user.ErrorResult.Error
	}

	_, err := s.handlePassword(&user)

	if err != nil {
		user.ErrorResult = ResultError(err)
		return User{ErrorResult: user.ErrorResult}, user.ErrorResult.Error
	}

	return s.repo.Register(user)
}

func (s *AuthoService) ChangePasswordService(user *User, newPassword string) (*User, error) {
	if user == nil || strings.TrimSpace(newPassword) == "" {
		return nil, fmt.Errorf("Invalid new password")
	}

	hashedPassword, err := s.HashPassword(newPassword)
	if err != nil {
		return nil, fmt.Errorf("cannot update password:: %v", err)
	}

	user.Password = hashedPassword

	return s.repo.ChangePassword(user, newPassword)
}

func (s *AuthoService) ValidatePassword(user *User, inputPassword string) (bool, error) {

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(inputPassword))

	if err != nil {
		return false, fmt.Errorf("Invalid password")
	}

	return s.repo.ValidatePassword(user, inputPassword)
}

func (s *AuthoService) ShowName(user *User) {
	if user == nil {
		fmt.Printf("Invalid user")
		return
	}

	fmt.Printf("This name is::: %s", user.Username)
}

func (s *AuthoService) isValidUserAccount(user *User) bool {
	if user == nil {
		fmt.Errorf("Empty user")
		return false
	}

	if strings.TrimSpace(user.Username) == "" {
		user.ErrorResult = ResultError(fmt.Errorf("Invalid username"))

		return user.ErrorResult.IsSuccess
	}

	user.ErrorResult = ResultSuccess()

	return user.ErrorResult.IsSuccess
}

func (s *AuthoService) handlePassword(user *User) (*User, error) {
	if user == nil || strings.TrimSpace(user.Password) == "" {
		return user, fmt.Errorf("Invalid user password")
	}

	hashedPassword, err := s.HashPassword(user.Password)
	if err != nil {
		return user, fmt.Errorf("Cannot process password")
	}

	user.Password = hashedPassword

	return user, nil
}

func (s *AuthoService) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}
