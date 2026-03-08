package main

import (
	v1 "2_Go/api/v1"
	"2_Go/api/v1/authApi"
	"2_Go/api/v1/routers"
	"2_Go/internal/auth"
	"2_Go/internal/category"
	"2_Go/internal/config"
	"2_Go/internal/document"
	"2_Go/internal/repo"
	rateLimit "2_Go/middleware"
	"2_Go/middleware/authen"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/hththai/ocr"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func main() {
	// Load config first so IsProduction() works correctly
	cfg := config.Load()

	// Initialize logging with correct environment
	AddLogService()

	db, err := repo.InitDB(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	log.Info("Database connection successful")

	// Start gin http
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		cfIP := param.Request.Header.Get("CF-Connecting-IP")
		ipField := param.ClientIP
		if cfIP != "" {
			ipField = fmt.Sprintf("%s | %s", cfIP, param.ClientIP)
		}
		return fmt.Sprintf("[GIN] %s | %3d | %13v | %s | %-7s %q\n",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			param.StatusCode,
			param.Latency,
			ipField,
			param.Method,
			param.Path,
		)
	}))

	// Trust the immediate upstream proxy so c.ClientIP() resolves X-Real-IP / X-Forwarded-For.
	// Replace with specific proxy CIDRs in production for stricter security.
	if err := r.SetTrustedProxies(nil); err != nil {
		log.Warnf("SetTrustedProxies: %v", err)
	}

	r.Use(rateLimit.RequestLogger(log))

	// Redis middleware (production only)
	if config.IsProduction() {
		AddRedisService(r, cfg)
	}

	// CORS config.
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.Origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	catRepo := category.NewCategoryRepository(db)
	catService := category.NewCategoryService(catRepo)

	docRepo := document.NewDocumentRepository(db, catRepo)
	docService := document.NewDocumentService(docRepo, catRepo)
	ocrService := ocr.NewService(cfg.Anthropic.APIKey)

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
		CatSvc:  catService,
		OcrSvc:  ocrService,
		Logger:  log,
	}

	routers.RegisterV1Routes(r, deps, log)

	// Test endpoint
	authGrouptest.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	log.Infof("Server starting on :%s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// AddLogService configures the application logger.
// config.Load() must be called before this function.
func AddLogService() {
	logPath := "./app/log/myapp.log"

	// Check existing file state for diagnostics (visible in docker logs)
	if info, err := os.Stat(logPath); err == nil {
		fmt.Printf("[log] existing log file found: %s (%d bytes)\n", logPath, info.Size())
	} else {
		fmt.Printf("[log] no existing log file at %s: %v\n", logPath, err)
	}

	// Open with O_APPEND — guaranteed never to truncate existing content.
	// Rotates at 100 MB, keeps 10 backup files.
	rw, err := newRotatingWriter(logPath, 100, 10)
	if err != nil {
		fmt.Printf("[log] failed to open log file, falling back to stdout: %v\n", err)
		log.SetOutput(os.Stdout)
	} else {
		log.SetOutput(io.MultiWriter(os.Stdout, rw))
	}

	log.SetLevel(logrus.DebugLevel)
	log.Info("******APPLICATION STARTED*******")
}

// Register Redis Services.
func AddRedisService(r *gin.Engine, cfg *config.Config) {
	rdb := redis.NewClient(&redis.Options{Addr: cfg.Redis.Addr})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Redis not reachable: %v", err)
	}

	log.Info("Redis connection successfully!")

	limiter := rateLimit.NewRedisRateLimiter(rdb, cfg.RateLimit.Requests, time.Minute)
	r.Use(rateLimit.RateLimitMiddleware(limiter))
}
