package auth

import (
	"codepix/bank-api/adapters/claimsctx"
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
	claims := claimsctx.GetClaims(ctx)
	IDString, _ := claims[BankIDKey].(string)
	ID, _ := uuid.Parse(IDString)
	return ID
}

func UnaryTokenValidator(config config.Config) grpc.UnaryServerInterceptor {
	validateToken := validateToken(config)

	return func(ctx context.Context, req any,
		info *grpc.UnaryServerInfo, next grpc.UnaryHandler) (any, error) {
		claims, err := validateToken(ctx)
		if err != nil {
			return nil, err
		}
		newCtx := claimsctx.AddClaims(ctx, claims)
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
		newCtx := claimsctx.AddClaims(stream.Context(), claims)

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
		tokenString, ok := md[cfg.MetadataKey]
		if !ok {
			return nil, status.Error(codes.Unauthenticated, fmt.Sprintf(
				"validate token: %s metadata not set", cfg.MetadataKey))
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
				return nil, status.Error(codes.Unauthenticated, "validate token: invalid JWT")
			}
		}
		return claims, nil
	}
}
