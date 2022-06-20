package interactor

import (
	"codepix/bank-api/account"

	"github.com/google/uuid"
)

type Interactor interface {
	Register(input RegisterInput) (*RegisterOutput, error)
}

type RegisterInput struct {
	Number    account.Number
	OwnerName string
	BankID    uuid.UUID
}
type RegisterOutput struct {
	Account account.Account
	ID      uuid.UUID
}
