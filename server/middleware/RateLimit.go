package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// Rate limit
var ctx = context.Background()

type RedisRateLimiter struct {
	client *redis.Client
	limit  int
	window time.Duration
}

func NewRedisRateLimiter(client *redis.Client, limit int, window time.Duration) *RedisRateLimiter {
	return &RedisRateLimiter{client: client, limit: limit, window: window}
}

func (r *RedisRateLimiter) Allow(key string) bool {
	now := time.Now().Unix()

	redisKey := fmt.Sprintf("rate:%s:%d", key, now/int64(r.window.Seconds()))

	count, err := r.client.Incr(ctx, redisKey).Result()

	if err != nil {
		return false
	}

	if count == 1 {
		r.client.Expire(ctx, redisKey, r.window)
	}

	return count <= int64(r.limit)
}

func RateLimitMiddleware(r *RedisRateLimiter) gin.HandlerFunc {

	return func(c *gin.Context) {
		userID := c.GetHeader("X-User-ID")

		if userID == "" {
			userID = c.ClientIP()
		}

		if !r.Allow(userID) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			c.Abort()
			return
		}

		c.Next()
	}
}
