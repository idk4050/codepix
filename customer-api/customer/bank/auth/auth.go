package auth

import (
	"codepix/customer-api/adapters/claimsctx"
	"codepix/customer-api/adapters/httputils"
	"codepix/customer-api/config"
	"codepix/customer-api/customer/bank/apikey"
	apikeyrepository "codepix/customer-api/customer/bank/apikey/repository"
	"codepix/customer-api/lib/repositories"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
)

const BankIDKey = "bank_id"

type Authenticate struct {
	APIKeySecret string `json:"api_key_secret" validate:"required"`
}
type AuthReply struct {
	Token  string        `json:"token"`
	Claims jwt.MapClaims `json:"claims"`
}

func AddClaims(apiKeyRepository apikeyrepository.Repository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			request := httputils.Body(r, Authenticate{})
			hash := apikey.HashSecret(request.APIKeySecret)

			ID, err := apiKeyRepository.FindBankID(hash)
			if err != nil {
				httputils.Error(w, r, err,
					httputils.Mapping{
						&repositories.NotFoundError{}: http.StatusUnauthorized,
					})
				return
			}
			claims := jwt.MapClaims{
				BankIDKey: ID.String(),
			}
			ctx := claimsctx.AddClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func CreateToken(config config.Config) http.HandlerFunc {
	cfg := config.BankAuth

	return func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		expirationTime := now.Add(cfg.MinutesUntilExpiration)

		claims := claimsctx.GetClaims(r.Context())
		claims["iat"] = now.Unix()
		claims["nbf"] = now.Unix()
		claims["exp"] = expirationTime.Unix()

		token := jwt.NewWithClaims(cfg.SigningMethod, claims)
		tokenString, err := token.SignedString(cfg.SigningKey)
		if err != nil {
			httputils.Error(w, r, err, httputils.Mapping{
				&repositories.NotFoundError{}: http.StatusUnauthorized,
			})
			return
		}
		reply := &AuthReply{
			Token:  tokenString,
			Claims: claims,
		}
		httputils.Json(w, reply, http.StatusOK)
	}
}
