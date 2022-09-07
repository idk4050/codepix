package service

import (
	"codepix/example-bank-api/adapters/httputils"
	"codepix/example-bank-api/customer/account"
	"codepix/example-bank-api/customer/account/repository"
	customerauth "codepix/example-bank-api/customer/auth"
	"net/http"

	"github.com/google/uuid"
)

type Service struct {
	Repository repository.Repository
}

type Registered struct {
	ID uuid.UUID `json:"id"`
}

func (s Service) Register(w http.ResponseWriter, r *http.Request) {
	customerID := customerauth.GetCustomerID(r.Context())

	account := account.Account{
		Number: account.GenerateNumber(),
	}
	ID, err := s.Repository.Add(account, customerID)
	if err != nil {
		httputils.Error(w, r, err)
		return
	}
	reply := Registered{*ID}
	httputils.Json(w, reply, http.StatusCreated)
}

type Find struct {
	ID uuid.UUID `param:"account-id"`
}
type FindResult struct {
	Number string `json:"number"`
}

func (s Service) Find(w http.ResponseWriter, r *http.Request) {
	params := httputils.Params(r, Find{})

	account, err := s.Repository.Find(params.ID)
	if err != nil {
		httputils.Error(w, r, err)
		return
	}
	result := FindResult{
		Number: account.Number,
	}
	httputils.Json(w, result, http.StatusOK)
}

type Remove struct {
	ID uuid.UUID `param:"account-id"`
}

func (s Service) Remove(w http.ResponseWriter, r *http.Request) {
	params := httputils.Params(r, Remove{})

	err := s.Repository.Remove(params.ID)
	if err != nil {
		httputils.Error(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type List struct {
	CustomerID uuid.UUID `param:"customer-id"`
}
type ListResult struct {
	Accounts []repository.AccountListItem `json:"accounts"`
}

func (s Service) List(w http.ResponseWriter, r *http.Request) {
	params := httputils.Params(r, List{})
	accounts, err := s.Repository.List(params.CustomerID)
	if err != nil {
		httputils.Error(w, r, err)
		return
	}
	result := ListResult{
		Accounts: accounts,
	}
	httputils.Json(w, result, http.StatusOK)
}
