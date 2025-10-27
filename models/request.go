package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type RequestType string
type RequestStatus string

const (
	RequestTypeCleaning    RequestType = "cleaning"
	RequestTypeMaintenance RequestType = "maintenance"

	RequestStatusActive    RequestStatus = "active"
	RequestStatusCompleted RequestStatus = "completed"
	RequestStatusCancelled RequestStatus = "cancelled"
)

type Request struct {
	bun.BaseModel `bun:"table:requests,alias:req"`

	ID          int            `bun:"id,pk,autoincrement" json:"id"`
	UserID      *uuid.UUID     `bun:"user_id,type:uuid" json:"user_id,omitempty"`
	RoomID      int            `bun:"room_id,notnull" json:"room_id"`
	Type        RequestType    `bun:"type,notnull" json:"type"`
	Status      RequestStatus  `bun:"status,default:'active'" json:"status"`
	Description *string        `bun:"description" json:"description,omitempty"`
	CreatedAt   time.Time      `bun:"created_at,nullzero,default:now()" json:"created_at"`
	UpdatedAt   time.Time      `bun:"updated_at,nullzero,default:now()" json:"updated_at"`

	// Relations
	User *User `bun:"rel:belongs-to,join:user_id=id" json:"user,omitempty"`
	Room *Room `bun:"rel:belongs-to,join:room_id=id" json:"room,omitempty"`
}
