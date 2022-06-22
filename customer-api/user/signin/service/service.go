package service

import (
	"codepix/customer-api/adapters/httputils"
	"codepix/customer-api/lib/repositories"
	"codepix/customer-api/user/signin/interactor"
	"net/http"
)

type Service struct {
	Interactor interactor.Interactor
}

type Start struct {
	Email string `json:"email" validate:"required,email,max=100" mod:"trim"`
}

func (s Service) Start(w http.ResponseWriter, r *http.Request) {
	body := httputils.Body(r, Start{})
	command := interactor.Start{
		Email: body.Email,
	}
	err := s.Interactor.Start(command)
	if err != nil {
		httputils.Error(w, r, err, httputils.Mapping{
			&repositories.NotFoundError{}: http.StatusCreated,
		})
		return
	}
	w.WriteHeader(http.StatusCreated)
}

type Finish struct {
	Token string `param:"token"`
}

func (s Service) Finish(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		next.ServeHTTP(w, r)
	})
}
