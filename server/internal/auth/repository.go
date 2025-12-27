package auth

import (
	"database/sql"
	"fmt"
)

type AuthRepository interface {
	Register(tx *sql.Tx, account Account) (*Account, error)
	ValidateUser(db *sql.DB, username string) (*Account, error)
}

type AuthRepositoryImpl struct {
	db *sql.DB
}

func NewAuthRepoImpl(db *sql.DB) AuthRepository {
	return &AuthRepositoryImpl{db: db}
}

// Insert into database new account.
// TODO: check the existing account
// TODO: check if needs to return account as result.
func (r *AuthRepositoryImpl) Register(tx *sql.Tx, account Account) (*Account, error) {

	// Check if the account username exist.
	row := tx.QueryRow(`SELECT COUNT(username) FROM account WHERE username = ?`, account.Username)

	var count int

	if err := row.Scan(&count); err != nil {
		return nil, fmt.Errorf("failed to check username: %w", err)
	}

	if count > 0 {
		return nil, fmt.Errorf("username exists")
	}

	_, err := tx.Exec(`INSERT INTO account (guid, username, password) VALUES (?,?,?)`, account.DefaultObj.GUID, account.Username, account.Password)

	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (r *AuthRepositoryImpl) ValidateUser(db *sql.DB, username string) (*Account, error) {

	row := db.QueryRow(`SELECT username, password from account where username=?`, username)

	var account Account
	if err := row.Scan(&account.Username, &account.Password); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}

		return nil, err
	}

	return &account, nil
}
