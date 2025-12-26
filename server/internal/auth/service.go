package auth

import (
	"2_Go/internal/obj"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	authRepo AuthRepository
}

func NewAuthService(authRepo AuthRepository) *AuthService {
	return &AuthService{authRepo: authRepo}
}

// Register account.
func (s *AuthService) Register(account Account) (*Account, error) {
	if err := account.Validate(); err != nil {
		return nil, err
	}

	_, err := s.handlePassword(&account)
	if err != nil {
		account.ErrorResult = obj.ResultError(err)
		return nil, account.ErrorResult.Error
	}
	return s.authRepo.Register(account)
}

// Change Password.
func (s *AuthService) ChangePasswordService(account Account, newPassword string) (*Account, error) {
	if err := account.Validate(); err != nil {
		return nil, fmt.Errorf("invalid account")
	}

	// Copy account
	updatedAccount := account
	updatedAccount.Password = newPassword

	// Validate new Password.
	if err := updatedAccount.Validate(); err != nil {
		return nil, fmt.Errorf("invalid Password")
	}

	if _, err := s.handlePassword(&updatedAccount); err != nil {
		return nil, fmt.Errorf("unsuccess change password")
	}

	return &updatedAccount, nil

}

// Ecrypt password
func (s *AuthService) handlePassword(account *Account) (*Account, error) {
	if err := account.Validate(); err != nil {
		return nil, fmt.Errorf("invalid account")
	}

	hashedPassword, err := s.hashPassword(account.Password)
	if err != nil {
		return nil, fmt.Errorf("cannot process password")
	}

	account.Password = hashedPassword

	return account, nil
}

func (s *AuthService) hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}
