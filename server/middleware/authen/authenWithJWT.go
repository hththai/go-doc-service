package authen

import (
	"2_Go/internal/config"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func getSecretKey() []byte {
	return []byte(config.Get().JWT.Secret)
}

// For testing purpose
func SayHello() error {
	fmt.Printf("hello")
	return nil
}

// Short token Access Token
// TODO: working on token id.
// func CreateAccessToken(username string) (string, string, error) {
// 	// Create an access token id.
// 	tokenId := uuid.NewString()

// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
// 		jwt.MapClaims{
// 			"username": username,
// 			"tokenId":  tokenId,
// 			"exp":      time.Now().Add(time.Minute * 60).Unix(),
// 			"iat":      time.Now().Unix(),
// 			"type":     "access",
// 		})

// 	tokenString, err := token.SignedString(getSecretKey())

// 	if err != nil {
// 		return "", "", err
// 	}

// 	return tokenString, tokenId, nil
// }

func CreateAccessToken(userId int) (string, string, error) {
	// Create an access token id.
	tokenId := uuid.NewString()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			// "username": username,
			"userId":  userId,
			"tokenId": tokenId,
			"exp":     time.Now().Add(time.Minute * 60).Unix(),
			"iat":     time.Now().Unix(),
			"type":    "access",
		})

	tokenString, err := token.SignedString(getSecretKey())

	if err != nil {
		return "", "", err
	}

	return tokenString, tokenId, nil
}

func VerifyToken(tokenString string) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return getSecretKey(), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims := token.Claims.(jwt.MapClaims)
	return &claims, nil
}

// Create refresh token.
func CreateRefreshToken(userId int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"userId": userId,
			"exp":    time.Now().Add(7 * 24 * time.Hour).Unix(),
			"type":   "refresh",
		})

	singedToken, err := token.SignedString(getSecretKey())

	if err != nil {
		return "", err
	}

	return singedToken, nil
}

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")

		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorize"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(header, "Bearer ")

		// Validate if it access_token.
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return getSecretKey(), nil
		})
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
		}

		claims := token.Claims.(jwt.MapClaims)

		if claims["type"] != "access" {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid access token"})
			return
		}

		// claims, err := VerifyToken(tokenString)

		// if err != nil {
		// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		// 	c.Abort()
		// 	return
		// }

		c.Set("username", (claims)["username"])

		c.Next()
	}
}

// Example of using cookies access token
func JWTAuthByCookies() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("access_token")
		if err != nil {
			c.JSON(401, gin.H{"error": "missing token"})
			c.Abort()
			return
		}

		claims, err := VerifyToken(token)
		if err != nil {
			c.JSON(401, gin.H{"error": "expired or invalid token"})
			c.Abort()
			return
		}

		// Handle userId to int.
		raw := (*claims)["userId"]
		userIdFloat, ok := raw.(float64)

		if !ok {
			c.JSON(401, gin.H{"error": "invalid userId"})
			c.Abort()
			return
		}

		userId := int(userIdFloat)

		// c.Set("userId", (*claims)["userId"])
		c.Set("userId", userId)
		c.Next()
	}
}

// IsValidUsername checks if the JWT username matches the request parameter username.
// Returns true if validation fails (user is unauthorized), false if valid.
func IsValidUsername(c *gin.Context) bool {
	jwtUser := c.GetString("username") // from Token
	reqUser := c.Param("username")     // from request

	if jwtUser != reqUser {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return true
	}
	return false
}
