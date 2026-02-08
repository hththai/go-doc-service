package main

import (
	v1 "2_Go/api/v1"
	"2_Go/api/v1/authApi"
	"2_Go/api/v1/routers"
	"2_Go/internal/auth"
	"2_Go/internal/document"
	config "2_Go/internal/repo"
	rateLimit "2_Go/middleware"
	"2_Go/middleware/authen"

	"2_Go/utils"
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var log = logrus.New()

func main() {
	// Initialize logging first
	AddLogService()

	dbCredential := config.LoadConfig()

	db, err := config.InitDB(dbCredential)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	log.Info("Database connection successful")

	// Start gin http
	r := gin.Default()

	// Redis middleware (production only)
	if utils.IsProduction() {
		AddRedisService(r)
	}

	// CORS config.
	corsOrigins := strings.Split(os.Getenv("CORS_ORIGINS"), ",")
	r.Use(cors.New(cors.Config{
		AllowOrigins:     corsOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	docRepo := document.NewDocumentRepository(db)
	docService := document.NewDocumentService(docRepo)
	// Register service.
	acctRepo := auth.NewAuthRepoImpl(db)
	acctSvc := auth.NewAuthService(acctRepo)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// Public auth handler for register and login
	publicAuthHandler := authApi.NewHandler(acctSvc, log)
	r.POST("/register", publicAuthHandler.HandleRegister)
	r.POST("/login", publicAuthHandler.HandleLogin)

	// Protected
	authGrouptest := r.Group("/v2/auth", authen.JWTAuth())

	deps := &v1.Dependencies{
		AuthSvc: acctSvc,
		DocSvc:  *docService,
		Logger:  log,
	}

	routers.RegisterV1Routes(r, deps, log)

	// Test endpoint
	authGrouptest.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	log.Info("Server starting on :8088")
	if err := r.Run(":8088"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// AddLogService configures the application logger.
func AddLogService() {
	log.SetOutput(&lumberjack.Logger{
		Filename:   "./app/log/myapp.log",
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     7,
		Compress:   true,
	})

	log.SetLevel(logrus.DebugLevel)

	log.Info("******APPLICATION STARTED*******")
}

// Register Redis Services.
func AddRedisService(r *gin.Engine) {
	rdb := redis.NewClient(&redis.Options{Addr: "redis-service:6379"})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Redis not reachable: %v", err)
	}

	log.Info("Redis connection successfully!")

	limiter := rateLimit.NewRedisRateLimiter(rdb, 10, time.Minute)
	r.Use(rateLimit.RateLimitMiddleware(limiter))
}
