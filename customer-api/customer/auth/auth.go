package auth

import (
	"codepix/customer-api/adapters/httputils"
	"codepix/customer-api/adapters/jwtclaims"
	bankrepository "codepix/customer-api/customer/bank/repository"
	"codepix/customer-api/customer/repository"
	"codepix/customer-api/lib/repositories"
	userauth "codepix/customer-api/user/auth"
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
			if err == nil {
				newClaims := jwt.MapClaims{
					CustomerIDKey: *ID,
				}
				ctx := jwtclaims.AddClaims(r.Context(), newClaims)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			var notFound *repositories.NotFoundError
			if errors.As(err, &notFound) {
				next.ServeHTTP(w, r)
				return
			}
			httputils.Error(w, r, err)
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

func ClaimedAndParamIDsMatch(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claimedID := GetCustomerID(r.Context())

		params := httprouter.ParamsFromContext(r.Context())
		requestedID := params.ByName("customer-id")
		if requestedID != claimedID.String() {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func CustomerOwnsParamBankID(bankRepository bankrepository.Repository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			customerID := GetCustomerID(r.Context())

			params := httprouter.ParamsFromContext(r.Context())
			bankID, _ := uuid.Parse(params.ByName("bank-id"))

			err := bankRepository.ExistsWithCustomerID(bankID, customerID)
			if err != nil {
				httputils.Error(w, r, err, httputils.Mapping{
					&repositories.NotFoundError{}: http.StatusForbidden,
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
