package repository

import (
	"codepix/bank-api/pixkey"

	"github.com/google/uuid"
)

type Repository interface {
	Add(pixKey pixkey.PixKey, accountID uuid.UUID) (*uuid.UUID, error)
	Find(ID uuid.UUID) (*pixkey.PixKey, *IDs, error)
	FindByKey(key pixkey.Key) (*pixkey.PixKey, *IDs, error)
}

type IDs struct {
	PixKeyID  uuid.UUID
	AccountID uuid.UUID
	BankID    uuid.UUID
}
