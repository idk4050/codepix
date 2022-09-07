package service

import (
	"codepix/example-bank-api/adapters/httputils"
	"codepix/example-bank-api/customer"
	"codepix/example-bank-api/customer/repository"
	"net/http"

	"github.com/google/uuid"
)

type Service struct {
	Repository repository.Repository
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
