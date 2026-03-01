package documentApi

import (
	"github.com/gin-gonic/gin"
)

func Register(rg *gin.RouterGroup, h *DocumentHandler) {
	g := rg.Group("/")

	g.POST("/upload", h.HandleUpload)
	g.GET("/purchases", h.HandleGetPurchases)
	g.POST("/purchases", h.HandleCreatePurchase)
	g.GET("/purchases/:id", h.HandleGetPurchaseById)
	g.GET("/purchases/:id/items", h.HandleGetPurchaseItems)
	g.PATCH("/purchases/:id", h.HandleUpdatePurchase)
	g.DELETE("/purchases/:id", h.HandleDeletePurchase)
	g.GET("/file/:id", h.HandleServeFile)
}
