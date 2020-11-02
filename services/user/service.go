package user

import "context"

// Service interface
type Service interface {
	Register(ctx context.Context, email, passwords string) (string, error)
	Login(ctx context.Context, email, passwords string) (string, error)
}
