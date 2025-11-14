package internal

import (
	"fmt"
	"time"
)

type Document struct {
	GUID       string    `json:"guid"`
	Id         string    `json:"id"`
	Title      string    `json:"title"`
	Extension  string    `json:"extension"`
	Location   string    `json:"location"`
	CreatedAt  time.Time `json:"createdAt"`
	ModifiedAt time.Time `json:"modifiedAt"`
}

func CreateNewDoc(title string) *Document {
	doc := Document{Title: title}
	return &doc
}

func PrintHello() {
	fmt.Println("Hello")

}
