package document

import (
	"fmt"
	"time"
)

type Document struct {
	GUID        string    `json:"guid"`
	Id          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Extension   string    `json:"extension"`
	FilePath    string    `json:"filepath"`
	FileSize    float32   `json:"filesize"`
	CreatedAt   time.Time `json:"createdAt"`
	ModifiedAt  time.Time `json:"modifiedAt"`
}

func PrintHello() {
	fmt.Println("Hello")
}

func CreateNewDoc(title string) *Document {
	doc := Document{Title: title}
	return &doc
}
