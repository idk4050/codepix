package service

import (
	"codepix/customer-api/adapters/httputils"
	customerauth "codepix/customer-api/customer/auth"
	"codepix/customer-api/customer/bank"
	apikeyrepository "codepix/customer-api/customer/bank/apikey/repository"
	"codepix/customer-api/customer/bank/interactor"
	"codepix/customer-api/customer/bank/repository"
	"net/http"

	"github.com/google/uuid"
)

type Service struct {
	Interactor       interactor.Interactor
	Repository       repository.Repository
	APIKeyRepository apikeyrepository.Repository
}

type Register struct {
	Code bank.Code `json:"code" validate:"required"`
	Name string    `json:"name" validate:"required,alpha,max=100" mod:"trim"`
}
type Registered struct {
	ID uuid.UUID `json:"id"`
}

func (s Service) Register(w http.ResponseWriter, r *http.Request) {
	customerID := customerauth.GetCustomerID(r.Context())
	input := register(
		httputils.Body(r, Register{}),
		customerID,
	)
	output, err := s.Interactor.Register(input)
	if err != nil {
		httputils.Error(w, r, err)
		return
	}
	reply := registered(*output)
	httputils.Json(w, reply, http.StatusCreated)
}

func register(body Register, customerID uuid.UUID) interactor.RegisterInput {
	return interactor.RegisterInput{
		Code:       body.Code,
		Name:       body.Name,
		CustomerID: customerID,
	}
}
func registered(output interactor.RegisterOutput) Registered {
	return Registered{
		ID: output.ID,
	}
}

type Find struct {
	ID uuid.UUID `param:"bank-id"`
}
type FindResult struct {
	Code bank.Code `json:"code"`
	Name string    `json:"name"`
}

func (s Service) Find(w http.ResponseWriter, r *http.Request) {
	params := httputils.Params(r, Find{})

	bank, err := s.Repository.Find(params.ID)
	if err != nil {
		httputils.Error(w, r, err)
		return
	}
	result := findResult(*bank)
	httputils.Json(w, result, http.StatusOK)
}

func findResult(bank bank.Bank) FindResult {
	return FindResult{
		Code: bank.Code,
		Name: bank.Name,
	}
}

type Remove struct {
	ID uuid.UUID `param:"bank-id"`
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

type ListAPIKeys struct {
	BankID uuid.UUID `param:"bank-id"`
}
type ListAPIKeysResult struct {
	APIKeys []apikeyrepository.APIKeyListItem `json:"api_keys"`
}

func (s Service) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	params := httputils.Params(r, ListAPIKeys{})

	apiKeys, err := s.APIKeyRepository.List(params.BankID)
	if err != nil {
		httputils.Error(w, r, err)
		return
	}
	result := ListAPIKeysResult{
		APIKeys: apiKeys,
	}
	httputils.Json(w, result, http.StatusOK)
}
