package commandhandler

import (
	"codepix/customer-api/customer/signup"
	"codepix/customer-api/customer/signup/interactor"
	"codepix/customer-api/customer/signup/repository"
)

type CommandHandler struct {
	Repository repository.Repository
}

var _ interactor.Interactor = CommandHandler{}

func (ch CommandHandler) Start(command interactor.Start) error {
	token, err := signup.GenerateToken()
	if err != nil {
		return err
	}
	signUp := signup.SignUp{
		Name:  command.Name,
		Email: command.Email,
		Token: token,
	}
	return ch.Repository.Start(signUp)
}

func (ch CommandHandler) Finish(command interactor.Finish) error {
	return ch.Repository.Finish(command.Token)
}
