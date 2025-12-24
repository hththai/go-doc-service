package auth

import (
	"database/sql"
)

type AuthRepository interface {
	Register(account Account) (*Account, error)
}

type AuthRepositoryImpl struct {
	db *sql.DB
}

func NewAuthRepoImpl(db *sql.DB) AuthRepository {
	return &AuthRepositoryImpl{db: db}
}

func (r *AuthRepositoryImpl) Register(account Account) (*Account, error) {
	return &account, nil
}
