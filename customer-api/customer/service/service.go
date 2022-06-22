package service

import (
	"codepix/customer-api/adapters/httputils"
	"codepix/customer-api/customer"
	bankrepository "codepix/customer-api/customer/bank/repository"
	"codepix/customer-api/customer/repository"
	"net/http"

	"github.com/google/uuid"
)

type Service struct {
	Repository     repository.Repository
	BankRepository bankrepository.Repository
}

type Find struct {
	ID uuid.UUID `param:"customer-id"`
}
type FindResult struct {
	Name string `json:"name"`
}

func (s Service) Find(w http.ResponseWriter, r *http.Request) {
	params := httputils.Params(r, Find{})

	customer, err := s.Repository.Find(params.ID)
	if err != nil {
		httputils.Error(w, r, err)
		return
	}
	result := findResult(*customer)
	httputils.Json(w, result, http.StatusOK)
}

func findResult(customer customer.Customer) FindResult {
	return FindResult{
		Name: customer.Name,
	}
}

type ListBanks struct {
	CustomerID uuid.UUID `param:"customer-id"`
}
type ListBanksResult struct {
	Banks []bankrepository.BankListItem `json:"banks"`
}

func (s Service) ListBanks(w http.ResponseWriter, r *http.Request) {
	params := httputils.Params(r, ListBanks{})

	banks, err := s.BankRepository.List(params.CustomerID)
	if err != nil {
		httputils.Error(w, r, err)
		return
	}
	result := ListBanksResult{
		Banks: banks,
	}
	httputils.Json(w, result, http.StatusOK)
}
