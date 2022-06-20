package bankapi

import (
	"context"

	"go.uber.org/zap"
)

type BankAPI struct {
}

func New(loggerImpl *zap.Logger) (*BankAPI, error) {
	bankAPI := &BankAPI{}
	return bankAPI, nil
}

func (api BankAPI) Start(ctx context.Context) error {
	return nil
}

func (api BankAPI) Stop() error {
	return nil
}
