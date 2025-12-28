package main

import (
	v1 "2_Go/api/v1"
	"2_Go/internal/auth"
	"2_Go/internal/document"
	config "2_Go/internal/repo"
	rateLimit "2_Go/middleware"
	"2_Go/utils"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Function to preview PDF
func previewPDF(c *gin.Context) {
	id := c.Query("id")
	token := c.Query("token")

	if token != "abc123" {
		c.String(http.StatusUnauthorized, "Unauthorized")
		return
	}

	filePath := filepath.Join("./filedata/0/0/", id+".pdf")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.String(http.StatusNotFound, "File not found")
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "inline; filename="+strconv.Quote(id+".pdf"))

	// Stream file.
	c.File(filePath)
}

var log = logrus.New()

func main() {

	dbCredential := config.LoadConfig()

	db, err := config.InitDB(dbCredential)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	fmt.Println("Database connection successfully!")
	// doc := internal.CreateNewDoc("Hello")
	// **************EXAMPLE LOG**************

	log.SetOutput(&lumberjack.Logger{
		Filename:   "./app/log/myapp.log",
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     7,
		Compress:   true,
	})

	log.SetLevel(logrus.DebugLevel)

	log.Info("******APPLICATION STARTED*******")
	r := gin.Default()

	//TODO: make a function, or interface.
	// Redis middleware
	if utils.IsProduction() {
		rdb := redis.NewClient(&redis.Options{Addr: "redis-service:6379"})
		err = rdb.Ping(context.Background()).Err()
		if err != nil {
			log.Fatalf("Redis not reachable: %v", err)
		}

		fmt.Println("Redis connection successfully!")

		limiter := rateLimit.NewRedisRateLimiter(rdb, 10, time.Minute)
		r.Use(rateLimit.RateLimitMiddleware(limiter))
		//*******
	}

	// // TODO: Review the order of middleware
	// r := gin.Default()
	// r.Use(rateLimit.RateLimitMiddleware(limiter))

	docRepo := document.NewDocumentRepository(db)
	docService := document.NewDocumentService(docRepo)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/preview", previewPDF)

	r.POST("/upload", func(c *gin.Context) {

		err = v1.UploadDocument(c, docService, db)

		if err != nil {
			log.Errorf("%s Error Upload: %s", c.ClientIP(), err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		log.Debugf("%s Upload Success", c.ClientIP())
		c.JSON(http.StatusOK, gin.H{"message": "Success"})

	})

	// Register service.
	acctRepo := auth.NewAuthRepoImpl(db)
	acctSvc := auth.NewAuthService(acctRepo)

	r.POST("/register", func(c *gin.Context) {
		err = v1.Register(c, acctSvc, db)

		if err != nil {
			log.Errorf("%s Error Register: %s", c.ClientIP(), err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		log.Debugf("%s register success", c.ClientIP())
		c.JSON(http.StatusOK, gin.H{"message": "Success"})
	})

	r.POST("/login", func(c *gin.Context) {
		err = v1.Login(c, acctSvc, db)

		if err != nil {
			log.Errorf("%s Error Register: %s", c.ClientIP(), err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		log.Debugf("%s register success", c.ClientIP())
		c.JSON(http.StatusOK, gin.H{"message": "Success"})
	})

	r.POST("/changepassword", func(c *gin.Context) {
		err = v1.ChangePassword(c, acctSvc, db)

		if err != nil {
			log.Errorf("%s Error Register: %s", c.ClientIP(), err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		log.Debugf("%s register success", c.ClientIP())
		c.JSON(http.StatusOK, gin.H{"message": "Success"})
	})

	r.Run(":8088")
}
