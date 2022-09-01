package auth

import (
	"codepix/customer-api/adapters/httputils"
	"codepix/customer-api/adapters/jwtclaims"
	"codepix/customer-api/config"
	"codepix/customer-api/customer/bank/apikey"
	apikeyrepository "codepix/customer-api/customer/bank/apikey/repository"
	"codepix/customer-api/lib/repositories"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
)

const BankIDKey = "bank_id"

func AddClaims(apiKeyRepository apikeyrepository.Repository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			APIKey, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
			hash := apikey.HashSecret(string(APIKey))
			ID, err := apiKeyRepository.FindBankID(hash)
			if err != nil {
				httputils.Error(w, r, err,
					httputils.Mapping{
						&repositories.NotFoundError{}: http.StatusUnauthorized,
					})
				return
			}
			claims := jwt.MapClaims{
				BankIDKey: *ID,
			}
			ctx := jwtclaims.AddClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func CreateToken(config config.Config) http.HandlerFunc {
	cfg := config.BankAuth

	return func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		expirationTime := now.Add(cfg.TimeUntilExpiration)

		claims := jwtclaims.GetClaims(r.Context())
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
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(tokenString))
	}
}
