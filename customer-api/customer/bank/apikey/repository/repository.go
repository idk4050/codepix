package repository

import (
	"codepix/customer-api/customer/bank/apikey"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	Add(apiKey apikey.APIKey, bankID uuid.UUID) (*uuid.UUID, error)
	Remove(ID uuid.UUID) error
	List(bankID uuid.UUID) ([]APIKeyListItem, error)
	FindBankID(hash apikey.Hash) (*uuid.UUID, error)
}

type APIKeyListItem struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
