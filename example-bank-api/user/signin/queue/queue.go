package queue

import "github.com/google/uuid"

const StartedStream = "signin_started"

type Started struct {
	Email  string    `json:"email"`
	Token  string    `json:"token"`
	UserID uuid.UUID `json:"user_id"`
}
