package commandhandler

import (
	userrepository "codepix/customer-api/user/repository"
	"codepix/customer-api/user/signin"
	"codepix/customer-api/user/signin/interactor"
	"codepix/customer-api/user/signin/repository"
)

type CommandHandler struct {
	Repository     repository.Repository
	UserRepository userrepository.Repository
}

var _ interactor.Interactor = CommandHandler{}

func (ch CommandHandler) Start(command interactor.Start) error {
	err := ch.UserRepository.Exists(command.Email)
	if err != nil {
		return err
	}
	token, err := signin.GenerateToken()
	if err != nil {
		return err
	}
	signIn := signin.SignIn{
		Email: command.Email,
		Token: token,
	}
	return ch.Repository.Start(signIn)
}

func (ch CommandHandler) Finish(command interactor.Finish) error {
	return ch.Repository.Finish(command.Token)
}
