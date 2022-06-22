package service

import (
	"codepix/customer-api/adapters/httputils"
	"codepix/customer-api/customer/bank/apikey"
	"codepix/customer-api/customer/bank/apikey/interactor"
	"codepix/customer-api/customer/bank/apikey/repository"
	"net/http"

	"github.com/google/uuid"
)

type Service struct {
	Interactor interactor.Interactor
	Repository repository.Repository
}

type Create struct {
	Name string `json:"name" validate:"required,alpha,max=100" mod:"trim"`
}
type CreateParams struct {
	BankID uuid.UUID `param:"bank-id"`
}
type Created struct {
	ID     uuid.UUID     `json:"id"`
	Secret apikey.Secret `json:"secret"`
}

func (s Service) Create(w http.ResponseWriter, r *http.Request) {
	input := create(
		httputils.Body(r, Create{}),
		httputils.Params(r, CreateParams{}),
	)
	output, err := s.Interactor.Create(input)
	if err != nil {
		httputils.Error(w, r, err)
		return
	}
	reply := created(*output)
	httputils.Json(w, reply, http.StatusCreated)
}

func create(body Create, params CreateParams) interactor.CreateInput {
	return interactor.CreateInput{
		Name:   body.Name,
		BankID: params.BankID,
	}
}
func created(output interactor.CreateOutput) Created {
	return Created{
		ID:     output.ID,
		Secret: output.APIKey.Secret,
	}
}

type Remove struct {
	ID uuid.UUID `param:"apikey-id"`
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
