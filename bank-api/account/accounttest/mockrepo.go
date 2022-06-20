package accounttest

import (
	"codepix/bank-api/account"
	"codepix/bank-api/account/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockRepo struct {
	mock.Mock
}

var _ repository.Repository = MockRepo{}

func (m MockRepo) Add(account account.Account, bankID uuid.UUID) (*uuid.UUID, error) {
	args := m.Called(account, bankID)
	return get[*uuid.UUID](args, 0), get[error](args, 1)
}

func (m MockRepo) Find(ID uuid.UUID) (*account.Account, *repository.IDs, error) {
	args := m.Called(ID)
	return get[*account.Account](args, 0), get[*repository.IDs](args, 1), get[error](args, 2)
}

func (m MockRepo) FindByNumber(number account.Number) (*account.Account, *repository.IDs, error) {
	args := m.Called(number)
	return get[*account.Account](args, 0), get[*repository.IDs](args, 1), get[error](args, 2)
}

func (m MockRepo) ExistsWithBankID(ID uuid.UUID, bankID uuid.UUID) error {
	args := m.Called(ID)
	return get[error](args, 0)
}

func get[T any](args mock.Arguments, index int) T {
	if args[index] == nil {
		return *new(T)
	}
	return args[index].(T)
}
