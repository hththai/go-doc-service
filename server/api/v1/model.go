package v1

import (
	"2_Go/internal/auth"
	"2_Go/internal/document"
	"database/sql"

	"github.com/sirupsen/logrus"
)

type Dependencies struct {
	AuthSvc auth.AuthService
	DocSvc  document.DocumentService
	DB      *sql.DB
	Logger  *logrus.Logger
	// Config     *config.Config
}
