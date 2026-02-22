package v1

import (
	"2_Go/internal/auth"
	"2_Go/internal/document"
	"github.com/hththai/ocr"

	"github.com/sirupsen/logrus"
)

type Dependencies struct {
	AuthSvc auth.AuthService
	DocSvc  document.DocumentService
	OcrSvc  *ocr.Service
	Logger  *logrus.Logger
	// Config     *config.Config
}
