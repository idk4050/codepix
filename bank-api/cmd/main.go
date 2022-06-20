package main

import (
	bankapi "codepix/bank-api"
	"codepix/bank-api/config"
	"context"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	ctx, cancelCreate := context.WithCancel(context.Background())

	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	config, err := config.New()
	if err != nil {
		logger.Fatal("failed to create config", zap.Error(err))
	}
	bankAPI, err := bankapi.New(ctx, logger, *config)
	if err != nil {
		cancelCreate()
		logger.Fatal("failed to create bank API", zap.Error(err))
	}

	ctx, cancelStart := context.WithCancel(context.Background())

	err = bankAPI.Start(ctx)
	if err != nil {
		logger.Error("bank API failed to start", zap.Error(err))
		stop(bankAPI, logger, cancelStart)
		return
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs

	stop(bankAPI, logger, cancelStart)
}

func stop(bankAPI *bankapi.BankAPI, logger *zap.Logger, cancelStart context.CancelFunc) {
	cancelStart()

	err := bankAPI.Stop()
	if err != nil {
		logger.Fatal("bank API failed to stop", zap.Error(err))
	}
}
