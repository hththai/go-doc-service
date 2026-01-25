package authApi

import (
	"github.com/gin-gonic/gin"
)

func Register(rg *gin.RouterGroup, h *AuthHandler) {
	g := rg.Group("/")

	g.GET("/test", h.GetPing)
	g.GET("/users/me", h.GetMe)
	g.POST("/refresh", h.HandleRefresh)
	g.POST("/users/:username/changepassword", h.HandleChangePassword)
	g.POST("/logout", h.HandleLogout)
}
