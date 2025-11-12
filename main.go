package main

import (
	"errors"
	item "example/golang/internal/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type todo struct {
	ID        string `json:"id"`
	Item      string `json:"title"`
	Completed bool   `json:"completed"`
}

var todos = []todo{
	{ID: "1", Item: "Clean Room", Completed: false},
	{ID: "2", Item: "Read Book", Completed: false},
	{ID: "3", Item: "Record Video", Completed: false},
}

func getTodos(context *gin.Context) {
	context.IndentedJSON(http.StatusOK, todos)
}

func addTodos(context *gin.Context) {
	var newTodo todo

	if err := context.BindJSON(&newTodo); err != nil {
		return
	}

	todos = append(todos, newTodo)

	context.IndentedJSON(http.StatusCreated, newTodo)

}

func getTodo(context *gin.Context) {
	id := context.Param("id")
	todo, err := getTodoById(id)

	if err != nil {
		context.IndentedJSON(http.StatusNotFound, gin.H{"message": "Todo not found"})
		return
	}

	context.IndentedJSON(http.StatusOK, todo)
}

func toggleTodoStatus(context *gin.Context) {
	id := context.Param("id")
	todo, err := getTodoById(id)

	if err != nil {
		context.IndentedJSON(http.StatusNotFound, gin.H{"message": "Todo not found"})
		return
	}

	todo.Completed = !todo.Completed

	context.IndentedJSON(http.StatusOK, todo)
}

func getTodoById(id string) (*todo, error) {
	for i, t := range todos {
		if t.ID == id {
			return &todos[i], nil
		}
	}

	return nil, errors.New("todo not found")
}

func main() {
	// fmt.Println("hello world")
	// router := gin.Default()
	// router.GET("/todos", getTodos)
	// router.GET("/todos/:id", getTodo)
	// router.PATCH("/todos/:id", toggleTodoStatus)
	// router.POST("/todos", addTodos)
	// router.Run("localhost:9090")

	// newItem := item.NewItem("Hello item")
	// // item.ShowItem(*newItem)
	// item.ShowItemAsJson(*newItem)
	// item.SaveToFileID(*newItem)

	items := []item.Item{
		{GUID: uuid.New().String(), Format: "pdf"},
		{GUID: uuid.New().String(), Format: "docx"},
		{GUID: uuid.New().String(), Format: "txt"},
		{GUID: uuid.New().String(), Format: "pdf"},
		{GUID: uuid.New().String(), Format: "docx"},
		{GUID: uuid.New().String(), Format: "txt"},
		{GUID: uuid.New().String(), Format: "pdf"},
		{GUID: uuid.New().String(), Format: "docx"},
		{GUID: uuid.New().String(), Format: "txt"},
		{GUID: uuid.New().String(), Format: "pdf"},
		{GUID: uuid.New().String(), Format: "docx"},
		{GUID: uuid.New().String(), Format: "txt"},
		{GUID: uuid.New().String(), Format: "pdf"},
		{GUID: uuid.New().String(), Format: "docx"},
		{GUID: uuid.New().String(), Format: "txt"},
		{GUID: uuid.New().String(), Format: "pdf"},
		{GUID: uuid.New().String(), Format: "docx"},
		{GUID: uuid.New().String(), Format: "txt"},
		{GUID: uuid.New().String(), Format: "pdf"},
		{GUID: uuid.New().String(), Format: "docx"},
		{GUID: uuid.New().String(), Format: "txt"},
		{GUID: uuid.New().String(), Format: "pdf"},
		{GUID: uuid.New().String(), Format: "docx"},
		{GUID: uuid.New().String(), Format: "txt"},
		{GUID: uuid.New().String(), Format: "pdf"},
		{GUID: uuid.New().String(), Format: "docx"},
		{GUID: uuid.New().String(), Format: "txt"},
		{GUID: uuid.New().String(), Format: "pdf"},
		{GUID: uuid.New().String(), Format: "docx"},
		{GUID: uuid.New().String(), Format: "txt"},
		{GUID: uuid.New().String(), Format: "pdf"},
		{GUID: uuid.New().String(), Format: "docx"},
		{GUID: uuid.New().String(), Format: "txt"},
		{GUID: uuid.New().String(), Format: "pdf"},
		{GUID: uuid.New().String(), Format: "docx"},
		{GUID: uuid.New().String(), Format: "txt"},
		{GUID: uuid.New().String(), Format: "pdf"},
		{GUID: uuid.New().String(), Format: "docx"},
		{GUID: uuid.New().String(), Format: "txt"},
		{GUID: uuid.New().String(), Format: "pdf"},
		{GUID: uuid.New().String(), Format: "docx"},
		{GUID: uuid.New().String(), Format: "txt"},
		{GUID: uuid.New().String(), Format: "pdf"},
		{GUID: uuid.New().String(), Format: "docx"},
		{GUID: uuid.New().String(), Format: "txt"},
		{GUID: uuid.New().String(), Format: "pdf"},
		{GUID: uuid.New().String(), Format: "docx"},
		{GUID: uuid.New().String(), Format: "txt"},
		{GUID: uuid.New().String(), Format: "pdf"},
		{GUID: uuid.New().String(), Format: "docx"},
		{GUID: uuid.New().String(), Format: "txt"},
		{GUID: uuid.New().String(), Format: "pdf"},
		{GUID: uuid.New().String(), Format: "docx"},
		{GUID: uuid.New().String(), Format: "txt"},
		{GUID: uuid.New().String(), Format: "pdf"},
		{GUID: uuid.New().String(), Format: "docx"},
		{GUID: uuid.New().String(), Format: "txt"},
		{GUID: uuid.New().String(), Format: "pdf"},
		{GUID: uuid.New().String(), Format: "docx"},
		{GUID: uuid.New().String(), Format: "txt"},
	}

	item.SaveCollectionFileToFileID(items)
	// item.ReadFileID()

}
