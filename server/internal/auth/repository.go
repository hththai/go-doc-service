package auth

import (
	"database/sql"
	"fmt"
)

type AuthRepository interface {
	Register(tx *sql.Tx, account Account) (*Account, error)
	ValidateUser(db *sql.DB, username string) (*Account, error)
	UpdatePassword(tx *sql.Tx, account Account) (*Account, error)
	GetUsrPassword(db *sql.DB, username string) (string, error)
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

// Get Hashpassword from user_name, and return only password.
// TODO: check and avoid the need of return account with user_name
func (r *AuthRepositoryImpl) GetUsrPassword(db *sql.DB, username string) (string, error) {
	row := db.QueryRow(`SELECT password FROM user WHERE user_name=?`, username)

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
