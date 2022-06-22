package interactor

import (
	"codepix/customer-api/customer/bank/apikey"

	"github.com/google/uuid"
)

type Interactor interface {
	Create(input CreateInput) (*CreateOutput, error)
}

type CreateInput struct {
	Name   string
	BankID uuid.UUID
}

type CreateOutput struct {
	APIKey apikey.APIKey
	ID     uuid.UUID
}
