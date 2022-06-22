package signup

import (
	"codepix/customer-api/adapters/encoding"
	"crypto/rand"
	"fmt"
)

type SignUp struct {
	Name  string
	Email string
	Token string
}

const TokenLength = 100

func GenerateToken() (string, error) {
	buff := make([]byte, TokenLength)

	_, err := rand.Read(buff)
	if err != nil {
		return "", fmt.Errorf("generate sign-up token: %w", err)
	}

	token := encoding.AlphaNumBase64.EncodeToString(buff)
	return token[:TokenLength], nil
}
