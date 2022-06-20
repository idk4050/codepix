package transactiontest

import (
	"codepix/bank-api/transaction/read/repository"
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockReadRepo struct {
	mock.Mock
}

var _ repository.Repository = MockReadRepo{}

func (m MockReadRepo) Find(ctx context.Context, ID uuid.UUID) (*repository.Transaction, error) {
	args := m.Called(ctx, ID)
	return get[*repository.Transaction](args, 0), get[error](args, 1)
}

func (m MockReadRepo) List(ctx context.Context, options repository.ListOptions,
) ([]repository.ListItem, error) {
	args := m.Called(ctx, options)
	return get[[]repository.ListItem](args, 0), get[error](args, 1)
}

func get[T any](args mock.Arguments, index int) T {
	if args[index] == nil {
		return *new(T)
	}
	return args[index].(T)
}
