package repository

import (
	"codepix/example-bank-api/customer/account"

	"github.com/google/uuid"
)

type Repository interface {
	Add(account account.Account, customerID uuid.UUID) (*uuid.UUID, error)
	Find(ID uuid.UUID) (*account.Account, error)
	Remove(ID uuid.UUID) error
	List(customerID uuid.UUID) ([]AccountListItem, error)
	ExistsWithCustomerID(ID uuid.UUID, customerID uuid.UUID) error
}

type AccountListItem struct {
	ID     uuid.UUID `json:"id"`
	Number string    `json:"number"`
}
