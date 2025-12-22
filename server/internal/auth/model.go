package auth

import (
	"2_Go/internal/obj"
)

type Account struct {
	Username   string         `json:"username"`
	Password   string         `json:"password"`
	ObjDefault obj.ObjDefault `json:"objDefault"`
}
