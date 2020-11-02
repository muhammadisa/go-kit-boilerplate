package user

import (
	"context"
	"time"

	uuid "github.com/satori/go.uuid"
)

// User model struct
type User struct {
	ID        uuid.UUID `json:"id,omitempty" db:"id"`
	Email     string    `json:"email"`
	Passwords string    `json:"passwords"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Repository interface for user
type Repository interface {
	Register(ctx context.Context, user User) error
	Login(ctx context.Context, email, passwords string) (*User, error)
}
