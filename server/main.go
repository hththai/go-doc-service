package main

import (
	v1 "2_Go/api/v1"
	"2_Go/internal/auth"
	"2_Go/internal/document"
	config "2_Go/internal/repo"
	rateLimit "2_Go/middleware"
	"2_Go/middleware/authen"

	"2_Go/utils"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
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

	// CORS config.
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return true
		},
	}))

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

	r.POST("/users/:username/changepassword", func(c *gin.Context) {

		err = v1.ChangePassword(c, acctSvc, db)

		if err != nil {
			log.Errorf("%s Error Register: %s", c.ClientIP(), err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		log.Debugf("%s register success", c.ClientIP())
		c.JSON(http.StatusOK, gin.H{"message": "Success"})
	})

	// Protected
	authGrouptest := r.Group("/v2/auth", authen.JWTAuth())
	authGroup := r.Group("/v1/auth", authen.JWTAuthByCookies())

	r.POST("/loginjwt", func(c *gin.Context) {
		token, err := v1.LoginJwt(c, acctSvc, db)

		if err != nil {
			log.Errorf("%s Error Register: %s", c.ClientIP(), err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		log.Debugf("%s register success", c.ClientIP())
		c.JSON(http.StatusOK, gin.H{"token": token})
	})

	// reset password endpoint.
	authGroup.POST("/users/:username/changepassword", func(c *gin.Context) {

		shouldReturn := IsValidUser(c)
		if shouldReturn {
			return
		}

		err = v1.ChangePassword(c, acctSvc, db)

		if err != nil {
			log.Errorf("%s Error Change: %s", c.ClientIP(), err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		log.Debugf("%s password updated success", c.ClientIP())
		c.JSON(http.StatusOK, gin.H{"message": "Success"})
	})

	// Refresh token endpoint.
	r.POST("/v1/auth/refresh", func(c *gin.Context) {
		err = v1.Refresh(c)

		if err != nil {
			log.Errorf("%s Error Change: %s", c.ClientIP(), err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		log.Debugf("%s password updated success", c.ClientIP())
		c.JSON(http.StatusOK, gin.H{"message": "Success"})
	})

	// Testing authorize.
	authGroup.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// validate access token.
	authGroup.GET("/me", func(c *gin.Context) {
		// TODO: should check IsValidUser

		err = v1.IsValidToken(c, acctSvc)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// logout
	authGroup.POST("/logout", func(c *gin.Context) {
		err = v1.Logout(c, acctSvc)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "logout"})
	})

	// Test
	authGrouptest.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	r.Run(":8088")
}

// validate if the user is matched the token.
func IsValidUser(c *gin.Context) bool {
	jwtUser := c.GetString("username") // from Token
	reqUser := c.Param("username")     // from request

	if jwtUser != reqUser {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return true
	}
	return false
}
