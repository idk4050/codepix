package auth

import (
	"codepix/example-bank-api/adapters/httputils"
	"codepix/example-bank-api/adapters/jwtclaims"
	"codepix/example-bank-api/config"
	"codepix/example-bank-api/lib/repositories"
	"codepix/example-bank-api/user/repository"
	signinrepository "codepix/example-bank-api/user/signin/repository"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

const UserIDKey = "user_id"

func GetUserID(ctx context.Context) uuid.UUID {
	claims := jwtclaims.GetClaims(ctx)
	return claims[UserIDKey].(uuid.UUID)
}

func AddClaims(repository repository.Repository, signInRepository signinrepository.Repository,
	tokenParam string,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			params := httprouter.ParamsFromContext(r.Context())
			token := params.ByName(tokenParam)

			_, IDs, err := signInRepository.Find(token)
			if err != nil {
				httputils.Error(w, r, fmt.Errorf("add user claims: find sign-in request: %w", err),
					httputils.Mapping{
						&repositories.NotFoundError{}: http.StatusUnauthorized,
					})
				return
			}
			claims := jwt.MapClaims{
				UserIDKey: IDs.UserID,
			}
			ctx := jwtclaims.AddClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func CreateToken(config config.Config) http.HandlerFunc {
	cfg := config.UserAuth

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
		http.SetCookie(w, &http.Cookie{
			Name:     cfg.CookieName,
			Value:    tokenString,
			Expires:  expirationTime,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})
		httputils.Json(w, claims, http.StatusOK)
	}
}

func ValidateToken(config config.Config) func(next http.Handler) http.Handler {
	cfg := config.UserAuth

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(cfg.CookieName)
			if err != nil {
				if err == http.ErrNoCookie {
					http.Error(w, fmt.Sprintf("validate token: %s cookie not set", cfg.CookieName),
						http.StatusUnauthorized)
					return
				}
				http.Error(w, fmt.Sprintf("validate token: %s cookie is invalid", cfg.CookieName),
					http.StatusUnauthorized)
				return
			}
			claims := jwt.MapClaims{}
			token, err := jwt.ParseWithClaims(cookie.Value, claims,
				func(token *jwt.Token) (interface{}, error) {
					return cfg.ValidationKey, nil
				},
			)
			if err != nil || !token.Valid {
				token, err := jwt.ParseWithClaims(cookie.Value, claims,
					func(token *jwt.Token) (interface{}, error) {
						return cfg.PreviousValidationKey, nil
					},
				)
				if err != nil || !token.Valid {
					http.Error(w, "validate token: invalid/expired token", http.StatusUnauthorized)
					return
				}
			}
			userIDString, _ := claims[UserIDKey].(string)
			userID, err := uuid.Parse(userIDString)
			if err != nil {
				http.Error(w, fmt.Sprintf("validate token: %s claim not set", UserIDKey), http.StatusUnauthorized)
			}
			claims[UserIDKey] = userID
			ctx := jwtclaims.AddClaims(r.Context(), claims)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
