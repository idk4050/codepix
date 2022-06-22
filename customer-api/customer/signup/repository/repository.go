package repository

import (
	"codepix/customer-api/customer/signup"
)

type Repository interface {
	Start(signUp signup.SignUp) error
	Finish(token string) error
}
