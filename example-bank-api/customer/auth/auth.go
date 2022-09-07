package auth

import (
	"codepix/example-bank-api/adapters/httputils"
	"codepix/example-bank-api/adapters/jwtclaims"
	"codepix/example-bank-api/customer/repository"
	"codepix/example-bank-api/lib/repositories"
	userauth "codepix/example-bank-api/user/auth"
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

const CustomerIDKey = "customer_id"

func GetCustomerID(ctx context.Context) uuid.UUID {
	claims := jwtclaims.GetClaims(ctx)
	return claims[CustomerIDKey].(uuid.UUID)
}

func AddClaims(repository repository.Repository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := userauth.GetUserID(r.Context())

			_, ID, err := repository.FindByUserID(userID)
			if err != nil {
				var notFound *repositories.NotFoundError
				if errors.As(err, &notFound) {
					next.ServeHTTP(w, r)
					return
				}
				httputils.Error(w, r, err)
			}
			newClaims := jwt.MapClaims{
				CustomerIDKey: *ID,
			}
			ctx := jwtclaims.AddClaims(r.Context(), newClaims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ValidateClaims(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := jwtclaims.GetClaims(r.Context())

		IDString, _ := claims[CustomerIDKey].(string)
		ID, err := uuid.Parse(IDString)
		if err != nil {
			http.Error(w, fmt.Sprintf("validate token: %s claim not set", CustomerIDKey), http.StatusUnauthorized)
		}
		claims[CustomerIDKey] = ID
		ctx := jwtclaims.AddClaims(r.Context(), claims)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func ClaimedAndParamIDsMatch(param string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			params := httprouter.ParamsFromContext(r.Context())
			requestedID := params.ByName(param)
			claimedID := GetCustomerID(r.Context())
			if requestedID != claimedID.String() {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
