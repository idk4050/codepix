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
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	config, err := config.New()
	if err != nil {
		logger.Fatal("failed to create config", zap.Error(err))
	}
	bankAPI, err := bankapi.New(*config, logger)
	if err != nil {
		logger.Fatal("failed to create bank API", zap.Error(err))
	}

	ctx, cancelStart := context.WithCancel(context.Background())

	logger.Info("starting bank API")
	err = bankAPI.Start(ctx)
	if err != nil {
		logger.Error("bank API failed to start", zap.Error(err))
		stop(bankAPI, logger, cancelStart)
		return
	}
	logger.Info("bank API started")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs

	stop(bankAPI, logger, cancelStart)
}

func stop(bankAPI *bankapi.BankAPI, logger *zap.Logger, cancelStart context.CancelFunc) {
	cancelStart()

	logger.Info("stopping bank API")
	err := bankAPI.Stop()
	if err != nil {
		logger.Fatal("bank API failed to stop", zap.Error(err))
	}
	logger.Info("bank API stopped")
}
