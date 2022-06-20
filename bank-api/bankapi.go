package bankapi

import (
	"codepix/bank-api/config"
	"context"

	"go.uber.org/zap"
)

type BankAPI struct {
	config config.Config
}

func New(config config.Config, loggerImpl *zap.Logger) (*BankAPI, error) {
	bankAPI := &BankAPI{
		config: config,
	}
	return bankAPI, nil
}

func (api BankAPI) Start(ctx context.Context) error {
	return nil
}

func (api BankAPI) Stop() error {
	return nil
}
