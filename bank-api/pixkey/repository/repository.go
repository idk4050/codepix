package repository

import (
	"codepix/bank-api/pixkey"

	"github.com/google/uuid"
)

type Repository interface {
	Add(pixKey pixkey.PixKey, accountID, bankID uuid.UUID) (*uuid.UUID, error)
	Find(ID uuid.UUID) (*pixkey.PixKey, *IDs, error)
	FindByKey(key pixkey.Key) (*pixkey.PixKey, *IDs, error)
	List(options ListOptions) ([]ListItem, error)
}

type IDs struct {
	PixKeyID  uuid.UUID
	AccountID uuid.UUID
	BankID    uuid.UUID
}

type ListItem struct {
	ID   uuid.UUID
	Type pixkey.Type
	Key  pixkey.Key
}

type ListOptions struct {
	AccountID uuid.UUID
	BankID    uuid.UUID
}
