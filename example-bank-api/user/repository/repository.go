package repository

import (
	"codepix/example-bank-api/user"

	"github.com/google/uuid"
)

type Repository interface {
	Add(user user.User) (*uuid.UUID, error)
	FindByEmail(email string) (*user.User, *uuid.UUID, error)
}
