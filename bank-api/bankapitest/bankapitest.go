package bankapitest

import (
	"codepix/bank-api/adapters/rpc"
	"codepix/bank-api/bank/auth"
	"codepix/bank-api/bank/authtest"
	"codepix/bank-api/config"
	"codepix/bank-api/lib/validation"
	"context"
	"net"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

var Config = func() config.Config {
	config, err := config.New()
	if err != nil {
		panic(err)
	}
	return *config
}()

var LoggerImpl = func() *zap.Logger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	return logger
}()

var Logger = func() logr.Logger {
	return zapr.NewLogger(LoggerImpl.WithOptions(
		zap.AddStacktrace(zapcore.DPanicLevel),
		zap.WithCaller(false),
	))
}()

var Server = func(validator *validation.Validator) (*grpc.Server, *grpc.ClientConn, func()) {
	grpcLogger := Logger.WithName("grpc")

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			rpc.UnaryLogger(grpcLogger),
			rpc.UnaryValidator(validator),
			auth.UnaryTokenValidator(Config),
		),
		grpc.ChainStreamInterceptor(
			rpc.StreamLogger(grpcLogger),
			rpc.StreamValidator(validator),
			auth.StreamTokenValidator(Config),
		),
	)
	listener := bufconn.Listen(1024 * 1024)

	client, err := grpc.DialContext(
		context.Background(),
		"",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return listener.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		panic(err)
	}
	serve := func() {
		go server.Serve(listener)
	}
	return server, client, serve
}

var AuthenticatedContext = authtest.AuthenticatedContext(Config)
