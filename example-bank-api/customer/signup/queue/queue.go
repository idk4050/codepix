package queue

const StartedStream = "signup_started"
const FinishedStream = "signup_finished"

type Started struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Token string `json:"token"`
}

type Finished struct {
	Token string `json:"token"`
}
