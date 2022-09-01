package interactor

type Interactor interface {
	Start(command Start) error
	Finish(command Finish) error
}

type Start struct {
	Name  string
	Email string
}

type Finish struct {
	Token string
}
