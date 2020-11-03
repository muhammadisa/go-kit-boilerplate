package implementation

import (
	"context"
	"errors"
	"time"

	"github.com/muhammadisa/go-kit-boilerplate/services/user"

	uuid "github.com/satori/go.uuid"

	"github.com/muhammadisa/go-kit-boilerplate/services/user/auth"
)

// userService struct
type userService struct {
	repository user.Repository
}

// NewService create instance of userService struct
func NewService(repo user.Repository) user.Service {
	return &userService{
		repository: repo,
	}
}

// Register logic function
func (service userService) Register(
	ctx context.Context,
	email, passwords string,
) (string, error) {
	hashedPassword, err := auth.HashPassword(passwords)
	if err != nil {
		return "", err
	}
	newUUID := uuid.NewV4()
	newUser := user.User{
		ID:        newUUID,
		Email:     email,
		Passwords: string(hashedPassword),
		CreatedAt: time.Now(),
	}
	if err := service.repository.Register(ctx, newUser); err != nil {
		return "", err
	}
	return "Success", nil
}

// Login logic function
func (service userService) Login(
	ctx context.Context,
	email, passwords string,
) (string, error) {
	selectedUser, err := service.repository.Login(ctx, email, passwords)
	if err != nil {
		return "", err
	}
	err = auth.VerifyPassword(selectedUser.Passwords, passwords)
	if err != nil {
		return "", errors.New("email or password is incorrect")
	}
	return "Success", nil
}
