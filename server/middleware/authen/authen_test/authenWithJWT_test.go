package authentest

import (
	"2_Go/middleware/authen"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

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
