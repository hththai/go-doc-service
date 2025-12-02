
package main

import (
    "net/http"
    "sync"
    "time"

    "github.com/gin-gonic/gin"
    "golang.org/x/time/rate"
)

type RateLimiter struct {
    mu       sync.Mutex
    limiters map[string]*rate.Limiter
}

func NewRateLimiter() *RateLimiter {
    return &RateLimiter{
        limiters: make(map[string]*rate.Limiter),
    }
}

func (r *RateLimiter) GetLimiter(key string) *rate.Limiter {
    r.mu.Lock()
    defer r.mu.Unlock()

    limiter, exists := r.limiters[key]
    if !exists {
        // 5 requests per minute
        limiter = rate.NewLimiter(rate.Every(time.Minute/5), 5)
        r.limiters[key] = limiter
    }
    return limiter
}

func RateLimitMiddleware(r *RateLimiter) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetHeader("X-User-ID") // or use JWT claims
        if userID == "" {
            userID = c.ClientIP() // fallback to IP
        }

        limiter := r.GetLimiter(userID)
        if !limiter.Allow() {
            c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
            c.Abort()
            return
        }
        c.Next()
    }
}

func main() {
    r := gin.Default()
    rateLimiter := NewRateLimiter()

    r.Use(RateLimitMiddleware(rateLimiter))

    r.GET("/api", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"message": "Success"})
    })

    r.Run(":8080")
}
