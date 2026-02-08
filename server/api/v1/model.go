package v1

import (
	"2_Go/internal/auth"
	"2_Go/internal/document"

	"github.com/sirupsen/logrus"
)

type Dependencies struct {
	AuthSvc auth.AuthService
	DocSvc  document.DocumentService
	Logger  *logrus.Logger
	// Config     *config.Config
}
