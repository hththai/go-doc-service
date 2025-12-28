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

func TestCreateToken(t *testing.T) {
	token, err := authen.CreateToken("hthai")

	if err != nil {
		t.Errorf("error")
	}

	t.Logf("token value is::: %v", token)
}

func TestVerifyToken(t *testing.T) {
	token, err := authen.CreateToken("hthai")
	if err != nil {
		t.Fatalf("error of creating token")
	}

	var claims *jwt.MapClaims
	claims, err = authen.VerifyToken(token)

	if err != nil {
		t.Errorf("Failed to have token")
	}

	t.Logf("value username is %v", (*claims)["username"])
}
