package auth

import (
	"2_Go/internal/obj"
	"2_Go/middleware/authen"
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"
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

	// Transactional wrappers
	RegisterWithTransaction(db *sql.DB, account Account) error
	ChangePasswordWithTransaction(db *sql.DB, userId int, currentPassword, newPassword string) error

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
		return fmt.Errorf("invalid account")
	}

	var tempAcct Account
	tempAcct.Password = crtPwd

	// 2. Compare to input.
	err = s.CheckPassword(&tempAcct, inputPwd)

	if err != nil {
		return fmt.Errorf("incorrect password")
	}
	return nil

}

// Validate current user and password.
func (s *authService) ValidateAccountByIdService(db *sql.DB, userId int, inputPwd string) error {

	// 1. Get current pwd.
	var crtPwd string
	crtPwd, err := s.authRepo.GetUsrPasswordById(db, userId)

	if err != nil {
		return fmt.Errorf("invalid account")
	}

	var tempAcct Account
	tempAcct.Password = crtPwd

	// 2. Compare to input.
	err = s.CheckPassword(&tempAcct, inputPwd)

	if err != nil {
		return fmt.Errorf("incorrect password")
	}
	return nil

}

// Service Login.
// func (s *AuthService) Login(db *sql.DB, username, password string) (*Account, error) {
// 	// step 1: fetch user.
// 	// account, err := s.authRepo.ValidateUser(db, username)
// 	var account Account

// 	// Fetch only password.
// 	pwd, err := s.authRepo.GetUsrPassword(db, username)

// 	if err != nil {
// 		return nil, fmt.Errorf("cannot retrieve account")
// 	}

// 	account.Password = pwd

// 	// Compare password.
// 	if err := s.CheckPassword(&account, password); err != nil {
// 		return nil, fmt.Errorf("incorrect password")
// 	}

// 	return &account, err
// }

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
		return nil, fmt.Errorf("incorrect password")
	}

	return &account, err
}

// Comparing password.
func (s *authService) CheckPassword(account *Account, inputPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(inputPassword))

	if err != nil {
		return fmt.Errorf("invalid password")
	}

	return nil

}

// Ecrypt password
// Handle password when creating an account. Checking all require fields such as username, password.
func (s *authService) handlePasswordAcctCreation(account *Account) (*Account, error) {
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

// RegisterWithTransaction handles the full registration flow with transaction.
func (s *authService) RegisterWithTransaction(db *sql.DB, account Account) error {
	if err := account.Validate(); err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		}
	}()

	_, err = s.Register(tx, account)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	return nil
}

// ChangePasswordWithTransaction handles password change with transaction.
func (s *authService) ChangePasswordWithTransaction(db *sql.DB, userId int, currentPassword, newPassword string) error {
	// Validate current password
	if err := s.ValidateAccountByIdService(db, userId, currentPassword); err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		}
	}()

	account := Account{UserId: userId, Password: currentPassword}
	_, err = s.ChangePasswordService(tx, account, newPassword)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	return nil
}
