package apikey

import (
	"codepix/customer-api/adapters/encoding"
	"crypto/rand"
	"crypto/sha512"
	"fmt"
)

type Secret string
type Hash []byte

type APIKey struct {
	Name   string
	Secret Secret
	Hash   Hash
}

const KeyLength = 100

func New(name string) (*APIKey, error) {
	buff := make([]byte, KeyLength)
	_, err := rand.Read(buff)
	if err != nil {
		return nil, fmt.Errorf("generate API key secret: %w", err)
	}
	secret := encoding.AlphaNumBase64.EncodeToString(buff)
	secret = secret[:KeyLength]

	hash := HashSecret(secret)

	return &APIKey{
		Name:   name,
		Secret: Secret(secret),
		Hash:   hash,
	}, nil
}

func HashSecret(secret string) Hash {
	hash := sha512.Sum512([]byte(secret))
	return hash[:]
}
