package interactor

import (
	"codepix/bank-api/pixkey"

	"github.com/google/uuid"
)

type Interactor interface {
	Register(input RegisterInput) (*RegisterOutput, error)
}

type RegisterInput struct {
	Type      pixkey.Type
	Key       pixkey.Key
	AccountID uuid.UUID
}
type RegisterOutput struct {
	PixKey pixkey.PixKey
	ID     uuid.UUID
}
