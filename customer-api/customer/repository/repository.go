package repository

import (
	"codepix/customer-api/customer"

	"github.com/google/uuid"
)

type Repository interface {
	Find(ID uuid.UUID) (*customer.Customer, error)
	FindByUserID(ID uuid.UUID) (*customer.Customer, *uuid.UUID, error)
}
