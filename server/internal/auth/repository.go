package auth

import (
	"database/sql"
	"fmt"
)

type AuthRepository interface {
	Register(tx *sql.Tx, account Account) (*Account, error)
	ValidateUser(db *sql.DB, username string) (*Account, error)
	UpdatePassword(tx *sql.Tx, account Account) (*Account, error)
	UpdatePasswordById(tx *sql.Tx, account Account) (*Account, error)
	// GetUsrPassword(db *sql.DB, username string) (string, error)
	GetUsrPassword(db *sql.DB, username string) (int, string, error)
	GetUsrPasswordById(db *sql.DB, userId int) (string, error)
}

type AuthRepositoryImpl struct {
	db *sql.DB
}

func NewAuthRepoImpl(db *sql.DB) AuthRepository {
	return &AuthRepositoryImpl{db: db}
}

// Insert into database new account.
// TODO: check if needs to return account as result.
func (r *AuthRepositoryImpl) Register(tx *sql.Tx, account Account) (*Account, error) {

	// Check if the account username exist.
	row := tx.QueryRow(`SELECT COUNT(user_name) FROM user WHERE user_name = ?`, account.Username)

	var count int

	if err := row.Scan(&count); err != nil {
		return nil, fmt.Errorf("failed to check username: %w", err)
	}

	if count > 0 {
		return nil, fmt.Errorf("username exists")
	}

	_, err := tx.Exec(`INSERT INTO user (guid, user_name, password) VALUES (?,?,?)`, account.DefaultObj.GUID, account.Username, account.Password)

	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (r *AuthRepositoryImpl) ValidateUser(db *sql.DB, username string) (*Account, error) {

	row := db.QueryRow(`SELECT user_name, password from user where user_name=?`, username)

	var account Account
	if err := row.Scan(&account.Username, &account.Password); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}

		return nil, err
	}

	return &account, nil
}

// Working: return value of Id and password.
func (r *AuthRepositoryImpl) GetUsrPassword(db *sql.DB, username string) (int, string, error) {
	row := db.QueryRow(`SELECT id, password FROM user WHERE user_name=?`, username)

	var id int
	var pwd string

	if err := row.Scan(&id, &pwd); err != nil {
		if err == sql.ErrNoRows {
			return -1, "", fmt.Errorf("no user found")
		}

		return -1, "", err
	}

	return id, pwd, nil
}

// Replacing GetUsrPassword
func (r *AuthRepositoryImpl) GetUsrPasswordById(db *sql.DB, userId int) (string, error) {
	row := db.QueryRow(`SELECT password FROM user WHERE id=?`, userId)

	var pwd string

	if err := row.Scan(&pwd); err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("no user found")
		}

		return "", err
	}

	return pwd, nil
}

// Change Password. Find the user and update password.
// Update based on username
func (r *AuthRepositoryImpl) UpdatePassword(tx *sql.Tx, account Account) (*Account, error) {
	_, err := tx.Exec(`
		UPDATE user SET password=? WHERE user_name=?
	`,
		account.Password, account.Username)

	if err != nil {
		return nil, err
	}

	return nil, nil
}

// Update password by userId
func (r *AuthRepositoryImpl) UpdatePasswordById(tx *sql.Tx, account Account) (*Account, error) {
	_, err := tx.Exec(`
		UPDATE user SET password=? WHERE id=?
	`,
		account.Password, account.UserId)

	if err != nil {
		return nil, err
	}

	return nil, nil
}

