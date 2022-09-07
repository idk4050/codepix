package repository

import (
	"codepix/example-bank-api/user/signin"

	"github.com/google/uuid"
)

type Repository interface {
	Add(signIn signin.SignIn, userID uuid.UUID) error
	Find(token string) (*signin.SignIn, *IDs, error)
	Remove(token string) error
}

type IDs struct {
	SignInID uuid.UUID
	UserID   uuid.UUID
}
