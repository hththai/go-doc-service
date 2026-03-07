package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RequestLogger returns a Gin middleware that logs each request with real client IP headers.
func RequestLogger(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		fields := logrus.Fields{
			"method":   c.Request.Method,
			"path":     c.Request.URL.Path,
			"status":   c.Writer.Status(),
			"latency":  time.Since(start).String(),
			"clientIP": c.ClientIP(),
		}

		if v := c.GetHeader("CF-Connecting-IP"); v != "" {
			fields["cf-connecting-ip"] = v
		}
		if v := c.GetHeader("X-Real-IP"); v != "" {
			fields["x-real-ip"] = v
		}
		if v := c.GetHeader("X-Forwarded-For"); v != "" {
			fields["x-forwarded-for"] = v
		}

		logger.WithFields(fields).Info("request")
	}
}
