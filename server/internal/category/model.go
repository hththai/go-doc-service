package category

import (
	"2_Go/internal/obj"
	"time"
)

// Category represents a user-owned label used to classify purchase documents.
type Category struct {
	GUID       string     `json:"guid"`
	Id         int        `json:"id"`
	UserId     int        `json:"userId"`
	Name       string     `json:"name"`
	Color      string     `json:"color"`
	Status     obj.Status `json:"status"`
	CreatedAt  time.Time  `json:"createdAt"`
	ModifiedAt time.Time  `json:"modifiedAt"`
}
