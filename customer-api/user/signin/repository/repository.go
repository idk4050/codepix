package repository

import (
	"codepix/customer-api/user/signin"
)

type Repository interface {
	Start(signIn signin.SignIn) error
	Find(token string) (*signin.SignIn, error)
	Finish(token string) error
}
