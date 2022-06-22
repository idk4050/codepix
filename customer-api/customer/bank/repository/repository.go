package repository

import (
	"codepix/customer-api/customer/bank"

	"github.com/google/uuid"
)

type Repository interface {
	Add(bank bank.Bank, customerID uuid.UUID) (*uuid.UUID, error)
	Find(ID uuid.UUID) (*bank.Bank, error)
	Remove(ID uuid.UUID) error
	List(customerID uuid.UUID) ([]BankListItem, error)
	ExistsWithCustomerID(ID uuid.UUID, customerID uuid.UUID) error
}

type BankListItem struct {
	ID   uuid.UUID `json:"id"`
	Code bank.Code `json:"code"`
	Name string    `json:"name"`
}
