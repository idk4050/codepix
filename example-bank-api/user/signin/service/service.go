package service

import (
	"codepix/example-bank-api/adapters/httputils"
	"codepix/example-bank-api/adapters/messagequeue"
	"codepix/example-bank-api/lib/repositories"
	userrepository "codepix/example-bank-api/user/repository"
	"codepix/example-bank-api/user/signin"
	"codepix/example-bank-api/user/signin/queue"
	"codepix/example-bank-api/user/signin/repository"
	"net/http"
)

type Service struct {
	MessageQueue   *messagequeue.MessageQueue
	Repository     repository.Repository
	UserRepository userrepository.Repository
}

type Start struct {
	Email string `json:"email" validate:"required,email,max=100" mod:"trim"`
}

func (s Service) Start(w http.ResponseWriter, r *http.Request) {
	body := httputils.Body(r, Start{})

	_, userID, err := s.UserRepository.FindByEmail(body.Email)
	if err != nil {
		httputils.Error(w, r, err)
		return
	}
	token, err := signin.GenerateToken()
	if err != nil {
		httputils.Error(w, r, err)
		return
	}
	message := queue.Started{
		Email:  body.Email,
		Token:  token,
		UserID: *userID,
	}
	err = s.MessageQueue.Write(r.Context(), message, []string{queue.StartedStream})
	if err != nil {
		httputils.Error(w, r, err)
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

		err := s.Repository.Remove(params.Token)
		if err != nil {
			httputils.Error(w, r, err, httputils.Mapping{
				&repositories.NotFoundError{}: http.StatusUnauthorized,
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}
