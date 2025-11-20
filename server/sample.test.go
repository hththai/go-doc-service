import (
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)


connString := "server=YOUR;database=YOUR;user id=YOUR;password=YOUR"

type Tester struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func main(){
  db, err := sql.Open("sqlserver", connString)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()
  
r := gin.Default()

	r.GET("/testers", func(c *gin.Context) {
		rows, err := db.Query("SELECT * from Tester")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var testers []Tester

		for rows.Next() {
			var tester Tester
			if err := rows.Scan(&tester.Id, &tester.Name); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			testers = append(testers, tester)
		}

		c.JSON(http.StatusOK, testers)
	})

	r.POST("/testers", func(c *gin.Context) {
		var newTester Tester
		if err := c.ShouldBindJSON(&newTester); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		query := "INSERT INTO Tester (name) VALUES (@p1)"
		_, err := db.Exec(query, newTester.Name)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "tester added"})
	})

	fmt.Println("Serever running....")
	r.Run(":8088")
}
