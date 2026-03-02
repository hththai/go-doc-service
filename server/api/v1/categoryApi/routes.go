package categoryApi

import "github.com/gin-gonic/gin"

func Register(rg *gin.RouterGroup, h *CategoryHandler) {
	g := rg.Group("/categories")
	g.GET("", h.HandleList)
	g.POST("", h.HandleCreate)
	g.PATCH("/:guid", h.HandleUpdate)
	g.DELETE("/:guid", h.HandleDelete)
}
