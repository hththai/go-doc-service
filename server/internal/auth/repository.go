package auth

import (
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type AuthRepository interface {
	Register(username string, password string) (Account, error)
}

type AuthRepositoryImpl struct {
	db *sql.DB
}

func NewAuthRepoImpl(db *sql.DB) AuthRepository {
	return &AuthRepositoryImpl{db: db}
}

func (r *AuthRepositoryImpl) Register(username string, password string) (Account, error) {

	hashPassword, err := hashPassword(password)

	if err != nil {
		return Account{}, fmt.Errorf("failed to create user")
	}

	newAccount := Account{
		Username: username,
		Password: hashPassword,
	}

	return newAccount, nil
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}
