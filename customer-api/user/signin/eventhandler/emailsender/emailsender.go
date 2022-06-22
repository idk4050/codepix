package emailsender

import (
	"codepix/customer-api/user/signin/eventhandler"
	"fmt"
)

type EmailSender struct {
}

var _ eventhandler.EventHandler = EmailSender{}

func (s EmailSender) Started(event eventhandler.Started) error {
	fmt.Println(event)
	return nil
}
