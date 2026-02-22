package document

import (
	"fmt"
	"time"
)

type Status int

const (
	StatusFailed  Status = -1 // -1
	StatusPending Status = 0
	StatusActive  Status = 1
)

var StatusLabels = map[Status]string{
	StatusFailed:  "failed",
	StatusPending: "pending",
	StatusActive:  "active",
}

type Document struct {
	GUID         string       `json:"guid"`
	Id           string       `json:"id"`
	UserId       int          `json:"userId"`
	Title        string       `json:"title"`
	Description  string       `json:"description"`
	Extension    string       `json:"extension"`
	FilePath     string       `json:"filepath"`
	FileSize     float64      `json:"filesize"`
	CreatedAt    time.Time    `json:"createdAt"`
	ModifiedAt   time.Time    `json:"modifiedAt"`
	Status       Status       `json:"status"`
	PurchaseInfo PurchaseInfo `json:"purchaseInfo"`
}

type PurchaseInfo struct {
	BuyAt    *time.Time `json:"buyAt"`
	BuyFrom  string     `json:"buyFrom"`
	BuyPrice string     `json:"buyPrice"`
}

func PrintHello() {
	fmt.Println("Hello")
}

func CreateNewDoc(title string) *Document {
	doc := Document{Title: title}
	return &doc
}
