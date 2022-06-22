package eventhandler

const Namespace = "signup"

type EventHandler interface {
	Started(event Started) error
	Finished(event Finished) error
}

type Started struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Token string `json:"token"`
}

func (Started) Namespace() string { return Namespace }
func (Started) Type() string      { return "started" }

type Finished struct {
	Email string `json:"email"`
}

func (Finished) Namespace() string { return Namespace }
func (Finished) Type() string      { return "finished" }
