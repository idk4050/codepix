package transactiontest

import (
	"context"

	"github.com/looplab/eventhorizon"
	"github.com/stretchr/testify/mock"
)

type MockCommandHandler struct {
	mock.Mock
}

var _ eventhorizon.CommandHandler = MockCommandHandler{}

func (m MockCommandHandler) HandleCommand(ctx context.Context, command eventhorizon.Command) error {
	args := m.Called(ctx, command)
	return get[error](args, 0)
}
