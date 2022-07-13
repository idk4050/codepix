package authtest

import (
	"codepix/bank-api/bank/auth"
	"codepix/bank-api/config"
	"context"
	"crypto/x509"
	"encoding/pem"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

func AuthenticatedContext(config config.Config) func(context.Context, uuid.UUID) context.Context {
	cfg := config.BankAuth

	return func(ctx context.Context, bankID uuid.UUID) context.Context {
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

		return metadata.AppendToOutgoingContext(ctx, cfg.MetadataKey, tokenString)
	}
}

var signingKeyString = []byte(`-----BEGIN PRIVATE KEY-----
MIICdwIBADANBgkqhkiG9w0BAQEFAASCAmEwggJdAgEAAoGBALhty5FBuBDxlxR6
TjeGRMzzEjsVHS+6TjuWBz+ylXf2x+R090TRjimPIS7a9QUAl7FyuTCmihzwSqn2
KPArPtTs8wkmXAz92EKdVDmGgc4VBOxdgF64vsmxfaCC0OJX/3yctKQ1fBHX2im/
kEUgNSTNkwlefv4tinpVFwWLwitDAgMBAAECgYEAhopKI7gej/W98gla//RqQlzc
Is+/T+8IXT2QIi6kDTSxE/8j10dL/xNT8Lt4XOLWxnNbl5sWTBAbV6ukp7fUiAEg
Q1k4moG/4EVzUgQyB4cZtT6y06PI9A2YV3AwBf0YldEaaqfQdSG24kcOMKt75+cL
juIgW6WQ2qUkqDbQCTECQQDZ/gOVbqQIuT1GH8Ed2q0+Aj3Y2WoQ9MdZqKiy7rm6
BsEF2aExWkprb1PSMr4wl68DVVVhvBR2/465pY1Bjor5AkEA2JWyJiEcSbIXX5Iz
aVflq6A8G9cN+scJxxrFW5gf3JHT6AfWNm7OSCk1j7iL9ztNWL84tCwuWAkynBhK
sqtbGwJAehwfF8rVWgmhuDE7dSS0nKKW0GzhTERBkwi2Dx1IrlrwLv28nK+uNkYz
VvCTtxaQs7ZOUKQRdqMq6PVCjjFxyQJBAKapstfyfLEdES1i9Jron3yNJhQKTeCf
Tx/esuYDzujNcrJHbYifha8zvtqkmVgbUy6qnzjOEq9+DGrfqoOIpucCQA1IDb+O
DKGoShF3EleSLXUmGNC1c0XcwdV9wq1d/wLX2peM3E6zd7KIAUJ1I3AiqBG5YxMq
5AyVBm4W1keTDDY=
-----END PRIVATE KEY-----`)

var signingKeyPem, _ = pem.Decode(signingKeyString)
var signingKey, _ = x509.ParsePKCS8PrivateKey(signingKeyPem.Bytes)
