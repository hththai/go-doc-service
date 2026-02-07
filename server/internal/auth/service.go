package auth

import (
	"2_Go/internal/obj"
	"2_Go/middleware/authen"
	"database/sql"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Error code.
var (
	ErrInvalidAccount    = errors.New("Account: Invalid Account")
	ErrIncorrectPassword = errors.New("Incorrect Password")
)

type AuthService interface {
	Register(tx *sql.Tx, account Account) (*Account, error)
	ChangePasswordService(tx *sql.Tx, account Account, newPassword string) (*Account, error)
	ValidateAccountService(db *sql.DB, username string, inputPwd string) error
	ValidateAccountByIdService(db *sql.DB, userId int, inputPwd string) error
	Login(db *sql.DB, username, password string) (*Account, error)
	CheckPassword(account *Account, inputPassword string) error

	// Token operations (no gin.Context dependency)
	CreateTokensForUser(userId int) (accessToken, tokenId, refreshToken string, err error)
	RefreshAccessToken(refreshToken string) (newAccessToken, newTokenId string, userId int, err error)
	ValidateAccessToken(accessToken string) (tokenId string, err error)

	// High-level operations (manage transactions internally)
	RegisterAccount(db *sql.DB, account Account) error
	ChangePassword(db *sql.DB, userId int, currentPassword, newPassword string) error

	handlePasswordAcctCreation(account *Account) (*Account, error)
	handlePasswordUpdate(account *Account) (*Account, error)
	hashPassword(password string) (string, error)
}

type authService struct {
	authRepo AuthRepository
}

func NewAuthService(authRepo AuthRepository) AuthService {
	return &authService{authRepo: authRepo}
}

// Register account.
func (s *authService) Register(tx *sql.Tx, account Account) (*Account, error) {
	if err := account.Validate(); err != nil {
		return nil, err
	}

	_, err := s.handlePasswordAcctCreation(&account)
	if err != nil {
		account.ErrorResult = obj.ResultError(err)
		return nil, account.ErrorResult.Error
	}
	return s.authRepo.Register(tx, account)
}

// RegisterAccount handles the full registration flow.
func (s *authService) RegisterAccount(db *sql.DB, account Account) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err = s.Register(tx, account); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	return nil
}

// Change Password.
func (s *authService) ChangePasswordService(tx *sql.Tx, account Account, newPassword string) (*Account, error) {
	// if err := account.Validate(); err != nil {
	// 	return nil, fmt.Errorf("invalid account")
	// }

	// Copy account
	updatedAccount := account
	updatedAccount.Password = newPassword

	// Validate new password
	if err := updatedAccount.ValidatePassword(); err != nil {
		return nil, fmt.Errorf("invalid password criteria")
	}

	if _, err := s.handlePasswordUpdate(&updatedAccount); err != nil {
		return nil, fmt.Errorf("unsuccess change password with Error %s", err)
	}

	// Add to repo.
	// _, err := s.authRepo.UpdatePassword(tx, updatedAccount)
	_, err := s.authRepo.UpdatePasswordById(tx, updatedAccount)
	if err != nil {
		return nil, err
	}

	return &updatedAccount, nil

}

// TODO: deleting replace by ValidateAccountByIdService
// Validate current username and password.
func (s *authService) ValidateAccountService(db *sql.DB, username string, inputPwd string) error {

	// 1. Get current pwd.
	var crtPwd string
	_, crtPwd, err := s.authRepo.GetUsrPassword(db, username)

	if err != nil {
		// return fmt.Errorf("invalid account")
		return ErrInvalidAccount
	}

	var tempAcct Account
	tempAcct.Password = crtPwd

	// 2. Compare to input.
	err = s.CheckPassword(&tempAcct, inputPwd)

	if err != nil {
		// return fmt.Errorf("incorrect password")
		return ErrIncorrectPassword
	}
	return nil

}

// Validate current user and password.
func (s *authService) ValidateAccountByIdService(db *sql.DB, userId int, inputPwd string) error {

	// 1. Get current pwd.
	var crtPwd string
	crtPwd, err := s.authRepo.GetUsrPasswordById(db, userId)

	if err != nil {
		// return fmt.Errorf("invalid account")
		return ErrInvalidAccount
	}

	var tempAcct Account
	tempAcct.Password = crtPwd

	// 2. Compare to input.
	err = s.CheckPassword(&tempAcct, inputPwd)

	if err != nil {
		// return fmt.Errorf("incorrect password")
		return ErrIncorrectPassword
	}
	return nil

}

// Working
func (s *authService) Login(db *sql.DB, username, password string) (*Account, error) {
	// step 1: fetch user.
	// account, err := s.authRepo.ValidateUser(db, username)
	var account Account

	// Fetch only password.
	id, pwd, err := s.authRepo.GetUsrPassword(db, username)

	if err != nil {
		return nil, fmt.Errorf("cannot retrieve account")
	}

	account.UserId = id
	account.Password = pwd

	// Compare password.
	if err := s.CheckPassword(&account, password); err != nil {
		// return nil, fmt.Errorf("incorrect password")
		return nil, ErrIncorrectPassword
	}

	return &account, err
}

// Comparing password.
func (s *authService) CheckPassword(account *Account, inputPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(inputPassword))

	if err != nil {
		// return fmt.Errorf("invalid password")
		return ErrIncorrectPassword
	}

	return nil

}

// Ecrypt password
// Handle password when creating an account. Checking all require fields such as username, password.
func (s *authService) handlePasswordAcctCreation(account *Account) (*Account, error) {
	if err := account.Validate(); err != nil {
		// return nil, fmt.Errorf("invalid account")
		return nil, ErrInvalidAccount
	}

	hashedPassword, err := s.hashPassword(account.Password)
	if err != nil {
		return nil, fmt.Errorf("cannot process password")
	}

	account.Password = hashedPassword

	return account, nil
}

// Check only password field needed.
func (s *authService) handlePasswordUpdate(account *Account) (*Account, error) {
	hashedPassword, err := s.hashPassword(account.Password)
	if err != nil {
		return nil, fmt.Errorf("cannot process password")
	}

	account.Password = hashedPassword

	return account, nil
}

func (s *authService) hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

// CreateTokensForUser generates access and refresh tokens for a user.
func (s *authService) CreateTokensForUser(userId int) (accessToken, tokenId, refreshToken string, err error) {
	accessToken, tokenId, err = authen.CreateAccessToken(userId)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create access token: %w", err)
	}

	refreshToken, err = authen.CreateRefreshToken(userId)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create refresh token: %w", err)
	}

	return accessToken, tokenId, refreshToken, nil
}

// RefreshAccessToken validates refresh token and issues new access token.
func (s *authService) RefreshAccessToken(refreshToken string) (newAccessToken, newTokenId string, userId int, err error) {
	claims, err := authen.VerifyToken(refreshToken)
	if err != nil {
		return "", "", 0, fmt.Errorf("invalid refresh token: %w", err)
	}

	raw := (*claims)["userId"]
	userIdFloat, ok := raw.(float64)
	if !ok {
		return "", "", 0, fmt.Errorf("invalid userId type in token")
	}
	userId = int(userIdFloat)

	newAccessToken, newTokenId, err = authen.CreateAccessToken(userId)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to create new access token: %w", err)
	}

	return newAccessToken, newTokenId, userId, nil
}

// ValidateAccessToken verifies an access token and returns the tokenId.
func (s *authService) ValidateAccessToken(accessToken string) (tokenId string, err error) {
	claims, err := authen.VerifyToken(accessToken)
	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}

	tokenId, ok := (*claims)["tokenId"].(string)
	if !ok {
		return "", fmt.Errorf("tokenId not found in token")
	}

	return tokenId, nil
}

// ChangePassword handles the full password change flow.
func (s *authService) ChangePassword(db *sql.DB, userId int, currentPassword, newPassword string) error {
	if err := s.ValidateAccountByIdService(db, userId, currentPassword); err != nil {
		return err
	}

	tempAccount := Account{Password: newPassword}
	if err := tempAccount.ValidatePassword(); err != nil {
		return fmt.Errorf("invalid password criteria")
	}

	hashedPassword, err := s.hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("cannot process password: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	account := Account{UserId: userId, Password: hashedPassword}
	if _, err = s.authRepo.UpdatePasswordById(tx, account); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	return nil
}
