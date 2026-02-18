package authentest

import (
	"2_Go/internal/config"
	"2_Go/middleware/authen"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func init() {
	config.Load()
}

func TestSayHello(t *testing.T) {
	err := authen.SayHello()

	if err != nil {
		t.Errorf("error")
	}

}

func TestCreateAccessToken(t *testing.T) {
	token, tokenId, err := authen.CreateAccessToken(524284)

	if err != nil {
		t.Errorf("error")
	}

	t.Logf("token value is::: %v", token)
	t.Logf("token id is ::: %v", tokenId)
}

func TestVerifyToken(t *testing.T) {
	token, _, err := authen.CreateAccessToken(524284)
	if err != nil {
		t.Fatalf("error of creating token")
	}

	var claims *jwt.MapClaims
	claims, err = authen.VerifyToken(token)

	if err != nil {
		t.Errorf("Failed to have token")
	}

	t.Logf("value tokenId is %v", (*claims)["tokenId"])
}

func TestCreateRefreshToken(t *testing.T) {
	token, err := authen.CreateRefreshToken(524284)
	if err != nil {
		t.Fatalf("error of creating refresh token %v", err)
	}

	t.Logf("refresh token is::: %v", token)
}

func TestConnectRedis(t *testing.T) {
	t.Logf("Hello")
}

func TestIsValidUsername_MatchingUsernames(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Set up the request with a username parameter
	c.Params = gin.Params{{Key: "username", Value: "testuser"}}
	// Set the JWT username in context (simulating what JWTAuth middleware does)
	c.Set("username", "testuser")

	result := authen.IsValidUsername(c)

	if result != false {
		t.Errorf("Expected false (valid), got true (invalid)")
	}

	if w.Code == http.StatusForbidden {
		t.Errorf("Expected no forbidden response, got %d", w.Code)
	}
}

func TestIsValidUsername_MismatchedUsernames(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Set up the request with a username parameter
	c.Params = gin.Params{{Key: "username", Value: "otheruser"}}
	// Set a different JWT username in context
	c.Set("username", "testuser")

	result := authen.IsValidUsername(c)

	if result != true {
		t.Errorf("Expected true (invalid), got false (valid)")
	}

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403 Forbidden, got %d", w.Code)
	}
}

func TestIsValidUsername_EmptyJWTUsername(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Set up the request with a username parameter
	c.Params = gin.Params{{Key: "username", Value: "testuser"}}
	// Don't set JWT username (simulating missing claim)

	result := authen.IsValidUsername(c)

	if result != true {
		t.Errorf("Expected true (invalid) when JWT username is empty, got false")
	}

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403 Forbidden, got %d", w.Code)
	}
}

// TestJWTAuthByCookiesValidToken verifies that a valid access_token cookie passes the middleware
// and sets userId in the gin context.
func TestJWTAuthByCookiesValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	expectedUserId := 99
	tokenString, _, err := authen.CreateAccessToken(expectedUserId)
	if err != nil {
		t.Fatalf("failed to create access token: %v", err)
	}

	r := gin.New()
	r.GET("/me", authen.JWTAuthByCookies(), func(c *gin.Context) {
		uid, exists := c.Get("userId")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "userId not set"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"userId": uid})
	})

	req, _ := http.NewRequest(http.MethodGet, "/me", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: tokenString})
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// TestJWTAuthByCookiesMissingCookie verifies that a request without the access_token cookie is rejected with 401.
func TestJWTAuthByCookiesMissingCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/me", authen.JWTAuthByCookies(), func(c *gin.Context) {
		c.JSON(http.StatusOK, nil)
	})

	req, _ := http.NewRequest(http.MethodGet, "/me", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// TestJWTAuthByCookiesInvalidToken verifies that a malformed or tampered token is rejected with 401.
func TestJWTAuthByCookiesInvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/me", authen.JWTAuthByCookies(), func(c *gin.Context) {
		c.JSON(http.StatusOK, nil)
	})

	req, _ := http.NewRequest(http.MethodGet, "/me", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "not.a.valid.jwt"})
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}
