package pixkeytest

import (
	"codepix/bank-api/pixkey/interactor"

	"github.com/stretchr/testify/mock"
)

type MockInteractor struct {
	mock.Mock
}

var _ interactor.Interactor = MockInteractor{}

func (m MockInteractor) Register(input interactor.RegisterInput) (*interactor.RegisterOutput, error) {
	args := m.Called(input)
	return get[*interactor.RegisterOutput](args, 0), get[error](args, 1)
}
