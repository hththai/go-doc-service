package documentApi

import (
	"github.com/gin-gonic/gin"
)

func Register(rg *gin.RouterGroup, h *DocumentHandler) {
	g := rg.Group("/")

	const purchaseRoute = "/purchases/:id"

	g.POST("/upload", h.HandleUpload)
	g.GET("/purchases", h.HandleGetPurchases)
	g.POST("/purchases", h.HandleCreatePurchase)
	g.GET(purchaseRoute, h.HandleGetPurchaseById)
	g.GET(purchaseRoute+"/items", h.HandleGetPurchaseItems)
	g.PATCH(purchaseRoute, h.HandleUpdatePurchase)
	g.DELETE(purchaseRoute, h.HandleDeletePurchase)
	g.GET("/file/:id", h.HandleServeFile)
}
