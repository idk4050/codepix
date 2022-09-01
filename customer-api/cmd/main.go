package main

import (
	customerapi "codepix/customer-api"
	"codepix/customer-api/config"
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
	customerAPI, err := customerapi.New(ctx, logger, *config)
	if err != nil {
		cancelCreate()
		logger.Fatal("failed to create customer API", zap.Error(err))
	}

	ctx, cancelStart := context.WithCancel(context.Background())

	err = customerAPI.Start(ctx)
	if err != nil {
		logger.Error("customer API failed to start", zap.Error(err))
		stop(customerAPI, logger, cancelStart)
		return
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs

	stop(customerAPI, logger, cancelStart)
}

func stop(customerAPI *customerapi.CustomerAPI, logger *zap.Logger, cancelStart context.CancelFunc) {
	cancelStart()

	err := customerAPI.Stop()
	if err != nil {
		logger.Fatal("customer API failed to stop", zap.Error(err))
	}
}
