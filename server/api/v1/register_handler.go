package v1

import (
	"2_Go/internal/auth"
	"database/sql"

	"github.com/gin-gonic/gin"
)

func Register(c *gin.Context, service *auth.AuthService, db *sql.DB) error {
	return nil
}
