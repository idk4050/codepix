package repository

import (
	"codepix/example-bank-api/customer/signup"
)

type Repository interface {
	Add(signUp signup.SignUp) error
	Remove(token string) error
}
