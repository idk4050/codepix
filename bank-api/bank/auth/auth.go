package auth

import (
	"codepix/bank-api/adapters/jwtclaims"
	"codepix/bank-api/adapters/rpc"
	"codepix/bank-api/config"
	"context"
	"fmt"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const BankIDKey = "bank_id"

func GetBankID(ctx context.Context) uuid.UUID {
	claims := jwtclaims.GetClaims(ctx)
	return claims[BankIDKey].(uuid.UUID)
}

func UnaryTokenValidator(config config.Config) grpc.UnaryServerInterceptor {
	validateToken := validateToken(config)

	return func(ctx context.Context, req any,
		info *grpc.UnaryServerInfo, next grpc.UnaryHandler) (any, error) {
		claims, err := validateToken(ctx)
		if err != nil {
			return nil, err
		}
		newCtx := jwtclaims.AddClaims(ctx, claims)
		return next(newCtx, req)
	}
}

func StreamTokenValidator(config config.Config) grpc.StreamServerInterceptor {
	validateToken := validateToken(config)

	return func(server any, stream grpc.ServerStream,
		info *grpc.StreamServerInfo, next grpc.StreamHandler) error {
		claims, err := validateToken(stream.Context())
		if err != nil {
			return err
		}
		newCtx := jwtclaims.AddClaims(stream.Context(), claims)

		streamWithCtx := &rpc.StreamWithCtx{ServerStream: stream, Ctx: newCtx}
		return next(server, streamWithCtx)
	}
}

func validateToken(config config.Config) func(context.Context) (jwt.MapClaims, error) {
	cfg := config.BankAuth

	return func(ctx context.Context) (jwt.MapClaims, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "validate token: no metadata received")
		}
		tokenString, ok := md["authorization"]
		if !ok {
			return nil, status.Error(codes.Unauthenticated, fmt.Sprintf(
				"validate token: '%s' metadata not set", "authorization"))
		}
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString[0], claims,
			func(token *jwt.Token) (interface{}, error) {
				return cfg.ValidationKey, nil
			},
		)
		if err != nil || !token.Valid {
			token, err := jwt.ParseWithClaims(tokenString[0], claims,
				func(token *jwt.Token) (interface{}, error) {
					return cfg.PreviousValidationKey, nil
				},
			)
			if err != nil || !token.Valid {
				return nil, status.Error(codes.Unauthenticated, "validate token: invalid/expired token")
			}
		}
		bankIDString, _ := claims[BankIDKey].(string)
		bankID, err := uuid.Parse(bankIDString)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "validate token: %s claim not set", BankIDKey)
		}
		claims[BankIDKey] = bankID
		return claims, nil
	}
}
