package bankapitest

import (
	"codepix/bank-api/adapters/rpc"
	"codepix/bank-api/bank/auth"
	"codepix/bank-api/config"
	"codepix/bank-api/lib/validation"
	"context"
	"crypto/x509"
	"encoding/pem"
	"net"
	"os"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
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

var PanicLogger = func() logr.Logger {
	return zapr.NewLogger(LoggerImpl.WithOptions(
		zap.AddStacktrace(zapcore.DebugLevel),
		zap.AddCallerSkip(3),
		zap.Fields(
			zap.StackSkip("stacktrace", 3),
		),
	))
}()

var Server = func(validator *validation.Validator) (*grpc.Server, *grpc.ClientConn, func()) {
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			rpc.UnaryPanicHandler(PanicLogger),
			rpc.UnaryLogger(Logger),
			auth.UnaryTokenValidator(Config),
			rpc.UnaryValidator(validator),
		),
		grpc.ChainStreamInterceptor(
			rpc.StreamPanicHandler(PanicLogger),
			rpc.StreamLogger(Logger),
			auth.StreamTokenValidator(Config),
			rpc.StreamValidator(validator),
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

func AuthenticatedContext(ctx context.Context, bankID uuid.UUID) context.Context {
	signingKeyPem, _ := pem.Decode([]byte(strings.ReplaceAll(os.Getenv("BANK_AUTH_SIGNING_KEY"), `\n`, "\n")))
	signingKey, _ := x509.ParsePKCS8PrivateKey(signingKeyPem.Bytes)

	now := time.Now()
	expirationTime := now.Add(time.Hour)

	claims := jwt.MapClaims{
		"iat":          now.Unix(),
		"nbf":          now.Unix(),
		"exp":          expirationTime.Unix(),
		auth.BankIDKey: bankID.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)
	tokenString, _ := token.SignedString(signingKey)

	return metadata.AppendToOutgoingContext(ctx, "authorization", tokenString)
}
