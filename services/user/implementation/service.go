package implementation

import (
	"context"
	"errors"
	"time"

	"github.com/muhammadisa/go-kit-boilerplate/services/user"

	uuid "github.com/satori/go.uuid"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/muhammadisa/go-kit-boilerplate/services/user/auth"
)

// userService struct
type userService struct {
	repository user.Repository
	logger     log.Logger
}

// NewService create instance of userService struct
func NewService(repo user.Repository, logger log.Logger) user.Service {
	return &userService{
		repository: repo,
		logger:     logger,
	}
}

// Register logic function
func (service userService) Register(
	ctx context.Context,
	email, passwords string,
) (string, error) {
	logger := log.With(service.logger, "method", "Register")

	hashedPassword, err := auth.HashPassword(passwords)
	if err != nil {
		level.Error(logger).Log("err", err)
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
		level.Error(logger).Log("err", err)
		return "", err
	}

	logger.Log("register user", newUUID.String())
	return "Success", nil
}

// Login logic function
func (service userService) Login(
	ctx context.Context,
	email, passwords string,
) (string, error) {
	logger := log.With(service.logger, "method", "Register")

	selectedUser, err := service.repository.Login(ctx, email, passwords)
	if err != nil {
		level.Error(logger).Log("err", err)
		return "", err
	}

	err = auth.VerifyPassword(selectedUser.Passwords, passwords)
	if err != nil {
		level.Error(logger).Log("err", err)
		return "", errors.New("email or password is incorrect")
	}

	logger.Log("login user", email)
	return "Success", nil
}
