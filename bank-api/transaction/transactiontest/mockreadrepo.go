package transactiontest

import (
	"codepix/bank-api/transaction/readrepository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockReadRepo struct {
	mock.Mock
}

var _ readrepository.ReadRepository = MockReadRepo{}

func (m MockReadRepo) Find(ID uuid.UUID) (*readrepository.Transaction, error) {
	args := m.Called(ID)
	return get[*readrepository.Transaction](args, 0), get[error](args, 1)
}

func get[T any](args mock.Arguments, index int) T {
	if args[index] == nil {
		return *new(T)
	}
	return args[index].(T)
}
