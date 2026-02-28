package document

import (
	"2_Go/internal/obj"
	"database/sql/driver"
	"fmt"
	"time"
)

// Date wraps time.Time and marshals to/from "DD/MM/YYYY" in JSON.
type Date struct{ time.Time }

func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.Format("02/01/2006") + `"`), nil
}

func (d Date) Value() (driver.Value, error) {
	return d.Format("2006-01-02"), nil
}

func (d *Date) UnmarshalJSON(b []byte) error {
	s := string(b)
	if s == "null" {
		return nil
	}
	// strip surrounding quotes
	s = s[1 : len(s)-1]
	t, err := time.Parse("02/01/2006", s)
	if err != nil {
		return err
	}
	d.Time = t
	return nil
}

type Document struct {
	GUID         string       `json:"guid"`
	Id           string       `json:"id"`
	UserId       int          `json:"userId"`
	Title        string       `json:"title"`
	FileName     string       `json:"fileName"`
	Description  string       `json:"description"`
	Extension    string       `json:"extension"`
	FilePath     string       `json:"filepath"`
	FileSize     float64      `json:"filesize"`
	CreatedAt    time.Time    `json:"createdAt"`
	ModifiedAt   time.Time    `json:"modifiedAt"`
	Status       obj.Status   `json:"status"`
	PurchaseInfo PurchaseInfo `json:"purchaseInfo"`
	Items        []Item       `json:"items"`
}

type PurchaseInfo struct {
	BuyAt    *Date  `json:"buyAt"`
	BuyFrom  string `json:"buyFrom"`
	BuyPrice string `json:"buyPrice"`
}

type Item struct {
	Name      string `json:"itemName"`
	Quantity  string `json:"itemQty"`
	UnitPrice string `json:"unitPrice"`
	SubTotal  string `json:"subTotal"`
}

func PrintHello() {
	fmt.Println("Hello")
}

func CreateNewDoc(title string) *Document {
	doc := Document{Title: title}
	return &doc
}
