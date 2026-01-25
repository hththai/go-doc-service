package documentApi

import (
	"github.com/gin-gonic/gin"
)

func Register(rg *gin.RouterGroup, h *DocumentHandler) {
	g := rg.Group("/")

	g.POST("/upload", h.HandleUpload)
}
