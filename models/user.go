package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID        uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	Name      string    `bun:"name,notnull" json:"name"`
	Email     string    `bun:"email,unique,notnull" json:"email"`
	Password  string    `bun:"password,notnull" json:"-"` // Don't expose password in JSON
	Block     *string   `bun:"block" json:"block,omitempty"`
	RoomName  *string   `bun:"room_name" json:"room_name,omitempty"`
	Phone     *string   `bun:"phone" json:"phone,omitempty"`
	CreatedAt time.Time `bun:"created_at,nullzero,default:now()" json:"created_at"`
}
