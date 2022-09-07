package account

import "github.com/google/uuid"

type Account struct {
	Number string
}

const NumberLength = 8

func GenerateNumber() string {
	randomID := uuid.NewString()
	return randomID[:NumberLength]
}
