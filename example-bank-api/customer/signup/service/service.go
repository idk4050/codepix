package service

import (
	"codepix/example-bank-api/adapters/httputils"
	"codepix/example-bank-api/adapters/messagequeue"
	"codepix/example-bank-api/customer/signup"
	"codepix/example-bank-api/customer/signup/queue"
	"codepix/example-bank-api/customer/signup/repository"
	"net/http"
)

type Service struct {
	MessageQueue *messagequeue.MessageQueue
	Repository   repository.Repository
}

type Start struct {
	Name  string `json:"name" validate:"required,alpha,max=100" mod:"trim"`
	Email string `json:"email" validate:"required,email,max=100" mod:"trim"`
}

func (s Service) Start(w http.ResponseWriter, r *http.Request) {
	body := httputils.Body(r, Start{})
	token, err := signup.GenerateToken()
	if err != nil {
		httputils.Error(w, r, err)
		return
	}
	started := queue.Started{
		Name:  body.Name,
		Email: body.Email,
		Token: token,
	}
	err = s.MessageQueue.Write(r.Context(), started, []string{queue.StartedStream})
	if err != nil {
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
	finished := queue.Finished{
		Token: params.Token,
	}
	err := s.MessageQueue.Write(r.Context(), finished, []string{queue.FinishedStream})
	if err != nil {
		httputils.Error(w, r, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
