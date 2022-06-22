package interactor

import (
	"codepix/customer-api/customer/bank"

	"github.com/google/uuid"
)

type Interactor interface {
	Register(input RegisterInput) (*RegisterOutput, error)
}

type RegisterInput struct {
	Code       bank.Code
	Name       string
	CustomerID uuid.UUID
}

type RegisterOutput struct {
	Bank bank.Bank
	ID   uuid.UUID
}
