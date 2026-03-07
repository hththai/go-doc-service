package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RequestLogger returns a Gin middleware that logs each request to the log file.
func RequestLogger(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		ip := c.GetHeader("CF-Connecting-IP")
		if ip == "" {
			ip = c.ClientIP()
		}

		logger.Infof("%d | %s | %s | %s %s",
			c.Writer.Status(),
			time.Since(start).Round(time.Microsecond),
			ip,
			c.Request.Method,
			c.Request.URL.Path,
		)
	}
}
