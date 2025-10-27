package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type RoomMember struct {
	bun.BaseModel `bun:"table:room_members,alias:rm"`

	RoomID   int       `bun:"room_id,pk" json:"room_id"`
	UserID   uuid.UUID `bun:"user_id,pk,type:uuid" json:"user_id"`
	JoinedAt time.Time `bun:"joined_at,nullzero,default:now()" json:"joined_at"`

	// Relations
	Room *Room `bun:"rel:belongs-to,join:room_id=id" json:"room,omitempty"`
	User *User `bun:"rel:belongs-to,join:user_id=id" json:"user,omitempty"`
}
