package routers

import (
	v1 "2_Go/api/v1"
	"2_Go/api/v1/authApi"
	"2_Go/api/v1/documentApi"
	"2_Go/middleware/authen"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func RegisterV1Routes(r *gin.Engine, deps *v1.Dependencies, logger *logrus.Logger) {
	apiv1 := r.Group("/v1/auth", authen.JWTAuthByCookies())

	authApi.Register(apiv1, authApi.NewHandler(deps.AuthSvc, logger))
	documentApi.Register(apiv1, documentApi.NewHandler(deps.AuthSvc, deps.DocSvc, logger))
}
