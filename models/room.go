package models

import (
	"time"

	"github.com/uptrace/bun"
)

type Room struct {
	bun.BaseModel `bun:"table:rooms,alias:r"`

	ID         int       `bun:"id,pk,autoincrement" json:"id"`
	Block      string    `bun:"block,notnull,unique:room_block" json:"block"`
	RoomNumber string    `bun:"room_number,notnull,unique:room_block" json:"room_number"`
	CreatedAt  time.Time `bun:"created_at,nullzero,default:now()" json:"created_at"`
}
