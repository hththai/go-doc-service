package maintenanceApi

import "github.com/gin-gonic/gin"

func Register(rg *gin.RouterGroup, h *MaintenanceHandler) {
	g := rg.Group("/maintenance")
	g.GET("/verify", h.HandleVerify)
	g.POST("/repair", h.HandleRepair)
}
