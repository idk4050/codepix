package main

import (
	examplebankapi "codepix/example-bank-api"
	"codepix/example-bank-api/config"
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
	api, err := examplebankapi.New(ctx, logger, *config)
	if err != nil {
		cancelCreate()
		logger.Fatal("failed to create Example Bank API", zap.Error(err))
	}

	ctx, cancelStart := context.WithCancel(context.Background())

	err = api.Start(ctx)
	if err != nil {
		logger.Error("Example Bank API failed to start", zap.Error(err))
		stop(api, logger, cancelStart)
		return
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs

	stop(api, logger, cancelStart)
}

func stop(api *examplebankapi.ExampleBankAPI, logger *zap.Logger, cancelStart context.CancelFunc) {
	cancelStart()

	err := api.Stop()
	if err != nil {
		logger.Fatal("Example Bank API failed to stop", zap.Error(err))
	}
}
