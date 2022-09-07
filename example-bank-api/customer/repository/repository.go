package repository

import (
	"codepix/example-bank-api/customer"

	"github.com/google/uuid"
)

type Repository interface {
	Find(ID uuid.UUID) (*customer.Customer, error)
	FindByUserID(ID uuid.UUID) (*customer.Customer, *uuid.UUID, error)
}
