package emailsender

import (
	"codepix/customer-api/customer/signup/eventhandler"
	"fmt"
)

type EmailSender struct {
}

var _ eventhandler.EventHandler = EmailSender{}

func (s EmailSender) Started(event eventhandler.Started) error {
	fmt.Println(event)
	return nil
}

func (s EmailSender) Finished(event eventhandler.Finished) error {
	fmt.Println(event)
	return nil
}
