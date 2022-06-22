package auth

import (
	"codepix/customer-api/adapters/claimsctx"
	"codepix/customer-api/adapters/httputils"
	bankrepository "codepix/customer-api/customer/bank/repository"
	"codepix/customer-api/customer/repository"
	"codepix/customer-api/lib/repositories"
	userauth "codepix/customer-api/user/auth"
	"context"
	"errors"
	"net/http"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

const CustomerIDKey = "customer_id"

func GetCustomerID(ctx context.Context) (uuid.UUID, bool) {
	claims := claimsctx.GetClaims(ctx)
	IDString, _ := claims[CustomerIDKey].(string)
	ID, err := uuid.Parse(IDString)
	return ID, err == nil
}

func AddClaims(repository repository.Repository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := userauth.GetUserID(r.Context())
			if !ok {
				httputils.Unauthorized(w)
				return
			}
			_, ID, err := repository.FindByUserID(userID)
			if err == nil {
				newClaims := jwt.MapClaims{
					CustomerIDKey: ID.String(),
				}
				ctx := claimsctx.AddClaims(r.Context(), newClaims)
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

func HasCustomerClaims(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, ok := GetCustomerID(r.Context())
		if !ok {
			httputils.Unauthorized(w)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func ClaimedAndParamIDsMatch(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claimedID, ok := GetCustomerID(r.Context())
		if !ok {
			httputils.Unauthorized(w)
			return
		}
		params := httprouter.ParamsFromContext(r.Context())
		requestedID := params.ByName("customer-id")
		if requestedID != claimedID.String() {
			httputils.Forbidden(w)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func CustomerOwnsParamBankID(bankRepository bankrepository.Repository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			customerID, ok := GetCustomerID(r.Context())
			if !ok {
				httputils.Unauthorized(w)
				return
			}
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
