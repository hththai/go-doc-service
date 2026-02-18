package authApi

import (
	"2_Go/middleware/authen"

	"github.com/gin-gonic/gin"
)

func Register(rg *gin.RouterGroup, h *AuthHandler) {
	g := rg.Group("/")

	g.GET("/test", h.GetPing)
	g.GET("/users/me", authen.JWTAuthByCookies(), h.GetMe)
	g.POST("/refresh", h.HandleRefresh)
	g.POST("/user/changepassword", h.HandleChangePassword)
	g.POST("/logout", h.HandleLogout)
}
