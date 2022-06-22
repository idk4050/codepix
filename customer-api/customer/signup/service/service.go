package service

import (
	"codepix/customer-api/adapters/httputils"
	"codepix/customer-api/customer/signup/interactor"
	"codepix/customer-api/lib/repositories"
	"net/http"
)

type Service struct {
	Interactor interactor.Interactor
}

type Start struct {
	Name  string `json:"name" validate:"required,alpha,max=100" mod:"trim"`
	Email string `json:"email" validate:"required,email,max=100" mod:"trim"`
}

func (s Service) Start(w http.ResponseWriter, r *http.Request) {
	body := httputils.Body(r, Start{})
	command := interactor.Start{
		Name:  body.Name,
		Email: body.Email,
	}
	if err := s.Interactor.Start(command); err != nil {
		httputils.Error(w, r, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

type Finish struct {
	Token string `param:"token"`
}

func (s Service) Finish(w http.ResponseWriter, r *http.Request) {
	params := httputils.Params(r, Finish{})
	command := interactor.Finish{
		Token: params.Token,
	}
	err := s.Interactor.Finish(command)
	if err != nil {
		httputils.Error(w, r, err, httputils.Mapping{
			&repositories.NotFoundError{}: http.StatusUnauthorized,
		})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
