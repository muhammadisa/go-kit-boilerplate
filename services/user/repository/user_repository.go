package repository

import (
	"context"
	"errors"

	"github.com/gocraft/dbr/v2"

	"github.com/go-kit/kit/log"
	"github.com/muhammadisa/go-kit-boilerplate/services/user"
)

// ErrRepo standard error for this repo
var ErrRepo = errors.New("Unable to handle repository request")

type repository struct {
	Session *dbr.Session
	logger  log.Logger
}

// NewUserRepository create instasnce of repo struct
func NewUserRepository(sess *dbr.Session, logger log.Logger) user.Repository {
	return &repository{
		Session: sess,
		logger:  logger,
	}
}

// Register database query logic
func (repo *repository) Register(
	ctx context.Context,
	user user.User,
) error {
	var err error

	_, err = repo.Session.InsertInto("users").
		Columns("id", "email", "passwords", "created_at").
		Record(user).
		Exec()
	if err != nil {
		return err
	}
	return nil
}

// Login database query logic
func (repo *repository) Login(
	ctx context.Context,
	email, passwords string,
) (*user.User, error) {
	var err error
	var selectedUser *user.User

	rowsAffected, err := repo.Session.Select("*").
		From("users").
		Where("email = ?", email).
		Load(&selectedUser)
	if rowsAffected == 0 {
		return nil, errors.New("User not found")
	}
	if err != nil {
		return nil, err
	}
	return selectedUser, nil
}
