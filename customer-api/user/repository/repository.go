package repository

import (
	"codepix/customer-api/user"

	"github.com/google/uuid"
)

type Repository interface {
	Add(user user.User) (*uuid.UUID, error)
	Find(email string) (*user.User, *uuid.UUID, error)
	Exists(email string) error
}
