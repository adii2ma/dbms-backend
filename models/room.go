package models

import (
	"time"

	"github.com/uptrace/bun"
)

type Room struct {
	bun.BaseModel `bun:"table:rooms,alias:r"`

	ID         int       `bun:"id,pk,autoincrement" json:"id"`
	RoomNumber string    `bun:"room_number,unique,notnull" json:"room_number"`
	CreatedAt  time.Time `bun:"created_at,nullzero,default:now()" json:"created_at"`
}
