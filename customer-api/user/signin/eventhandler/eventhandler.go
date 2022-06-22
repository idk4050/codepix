package eventhandler

const Namespace = "signin"

type EventHandler interface {
	Started(event Started) error
}

type Started struct {
	Email string `json:"email"`
	Token string `json:"token"`
}

func (Started) Namespace() string { return Namespace }
func (Started) Type() string      { return "started" }
