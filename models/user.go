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
	Phone     *string   `bun:"phone" json:"phone,omitempty"`
	CreatedAt time.Time `bun:"created_at,nullzero,default:now()" json:"created_at"`
}
