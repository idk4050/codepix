package pixkeytest

import (
	"codepix/bank-api/pixkey"
	"codepix/bank-api/pixkey/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockRepo struct {
	mock.Mock
}

var _ repository.Repository = MockRepo{}

func (m MockRepo) Add(pixKey pixkey.PixKey, accountID, bankID uuid.UUID) (*uuid.UUID, error) {
	args := m.Called(pixKey, accountID)
	return get[*uuid.UUID](args, 0), get[error](args, 1)
}
func (m MockRepo) Find(ID uuid.UUID) (*pixkey.PixKey, *repository.IDs, error) {
	args := m.Called(ID)
	return get[*pixkey.PixKey](args, 0), get[*repository.IDs](args, 1), get[error](args, 2)
}
func (m MockRepo) FindByKey(key pixkey.Key) (*pixkey.PixKey, *repository.IDs, error) {
	args := m.Called(key)
	return get[*pixkey.PixKey](args, 0), get[*repository.IDs](args, 1), get[error](args, 2)
}
func (m MockRepo) List(options repository.ListOptions) ([]repository.ListItem, error) {
	args := m.Called(options)
	return get[[]repository.ListItem](args, 0), get[error](args, 1)
}

func get[T any](args mock.Arguments, index int) T {
	if args[index] == nil {
		return *new(T)
	}
	return args[index].(T)
}
