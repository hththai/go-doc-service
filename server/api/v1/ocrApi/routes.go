package ocrApi

import "github.com/gin-gonic/gin"

func Register(rg *gin.RouterGroup, h *Handler) {
	rg.POST("/ocr", h.HandleOCR)
}
