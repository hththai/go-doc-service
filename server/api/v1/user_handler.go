package api

import (
	"fmt"
	"net/http"
	"sample-module/internal/user"

	"github.com/gin-gonic/gin"
)

func GetAllUser(c *gin.Context, u *user.UserService) (string, error) {
	result, err := u.Repo.FindAll()

	if err != nil {
		fmt.Printf("error %s", err)
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})

		return "", err
	}

	return result, nil

}
