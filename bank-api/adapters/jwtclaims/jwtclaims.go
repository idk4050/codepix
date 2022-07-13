package jwtclaims

import (
	"context"

	"github.com/golang-jwt/jwt"
)

type key string

const claimsKey key = "claims"

func AddClaims(ctx context.Context, newClaims jwt.MapClaims) context.Context {
	claims := GetClaims(ctx)
	for k, v := range newClaims {
		claims[k] = v
	}
	return context.WithValue(ctx, claimsKey, claims)
}

func GetClaims(ctx context.Context) jwt.MapClaims {
	claims, ok := ctx.Value(claimsKey).(jwt.MapClaims)
	if !ok {
		return jwt.MapClaims{}
	}
	return claims
}
