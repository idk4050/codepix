package repository

import (
	"codepix/bank-api/account"

	"github.com/google/uuid"
)

type Repository interface {
	Add(account account.Account, bankID uuid.UUID) (*uuid.UUID, error)
	Find(ID uuid.UUID) (*account.Account, *IDs, error)
	FindByNumber(number account.Number) (*account.Account, *IDs, error)
	ExistsWithBankID(ID uuid.UUID, bankID uuid.UUID) error
}

type IDs struct {
	AccountID uuid.UUID
	BankID    uuid.UUID
}
