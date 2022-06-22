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
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	config, err := config.New()
	if err != nil {
		logger.Fatal("failed to create config", zap.Error(err))
	}
	customerAPI, err := customerapi.New(*config, logger)
	if err != nil {
		logger.Fatal("failed to create customer API", zap.Error(err))
	}

	ctx, cancelStart := context.WithCancel(context.Background())

	logger.Info("starting customer API")
	err = customerAPI.Start(ctx)
	if err != nil {
		logger.Error("customer API failed to start", zap.Error(err))
		stop(customerAPI, logger, cancelStart)
		return
	}
	logger.Info("customer API started")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs

	stop(customerAPI, logger, cancelStart)
}

func stop(customerAPI *customerapi.CustomerAPI, logger *zap.Logger, cancelStart context.CancelFunc) {
	cancelStart()

	logger.Info("stopping customer API")
	err := customerAPI.Stop()
	if err != nil {
		logger.Fatal("customer API failed to stop", zap.Error(err))
	}
	logger.Info("customer API stopped")
}
