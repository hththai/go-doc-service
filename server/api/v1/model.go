package v1

import (
	"2_Go/internal/auth"
	"2_Go/internal/category"
	"2_Go/internal/document"
	"2_Go/internal/maintenance"
	"github.com/hththai/ocr"

	"github.com/sirupsen/logrus"
)

type Dependencies struct {
	AuthSvc auth.AuthService
	DocSvc  document.DocumentService
	CatSvc  *category.CategoryService
	MntSvc  *maintenance.MaintenanceService
	OcrSvc  *ocr.Service
	Logger  *logrus.Logger
	// Config     *config.Config
}
