package auth

import (
	"database/sql"
)

type AuthRepository interface {
	Register(tx *sql.Tx, account Account) (*Account, error)
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

	_, err := tx.Exec(`INSERT INTO account (guid, username, password) VALUES (?,?,?)`, account.DefaultObj.GUID, account.Username, account.Password)

	if err != nil {
		return nil, err
	}

	return nil, nil
}
