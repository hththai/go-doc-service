package authentest

import (
	"2_Go/middleware/authen"
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

func TestSayHello(t *testing.T) {
	err := authen.SayHello()

	if err != nil {
		t.Errorf("error")
	}

}

func TestCreateAccessToken(t *testing.T) {
	token, tokenId, err := authen.CreateAccessToken("hthai")

	if err != nil {
		t.Errorf("error")
	}

	t.Logf("token value is::: %v", token)
	t.Logf("token id is ::: %v", tokenId)
}

func TestVerifyToken(t *testing.T) {
	token, _, err := authen.CreateAccessToken("hthai")
	if err != nil {
		t.Fatalf("error of creating token")
	}

	var claims *jwt.MapClaims
	claims, err = authen.VerifyToken(token)

	if err != nil {
		t.Errorf("Failed to have token")
	}

	t.Logf("value username is %v", (*claims)["tokenId"])
}

func TestCreateRefreshToken(t *testing.T) {
	token, err := authen.CreateRefreshToken("hthai")
	if err != nil {
		t.Fatalf("error of creating refresh token %v", err)
	}

	t.Logf("refresh token is::: %v", token)
}

func TestConnectRedis_Integration(t *testing.T) {
	ctx := context.Background()

	// 1. Define the Redis container configuration using Testcontainers
	req := testcontainers.ContainerRequest{
		Image:        "redis:latest", // Use the official Redis image from Docker Hub
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp").WithStartupTimeout(2 * time.Minute),
	}

	// 2. Start the container
	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	// 3. Clean up the container after the test finishes
	defer func() {
		err := redisContainer.Terminate(ctx)
		require.NoError(t, err)
	}()

	// 4. Get the dynamically assigned host and port
	host, err := redisContainer.Host(ctx)
	require.NoError(t, err)
	port, err := redisContainer.MappedPort(ctx, "6379")
	require.NoError(t, err)

	redisAddr := host + ":" + port.Port()

	// 5. Run the actual function to test
	client, err := ConnectRedis(ctx, redisAddr)
	require.NoError(t, err, "ConnectRedis should not return an error")
	require.NotNil(t, client, "Client should not be nil")

	// 6. Optional: Test the functionality of the client
	key := "testkey"
	value := "testvalue"
	result, err := SetAndGet(ctx, client, key, value)
	require.NoError(t, err)
	require.Equal(t, value, result)
}
