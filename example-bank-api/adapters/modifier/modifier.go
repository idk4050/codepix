package modifier

import (
	"context"

	"github.com/go-playground/mold/v4/modifiers"
)

var MainModifier = modifiers.New()

// Mold applies transformations against struct s.
func Mold(s any) error {
	return MainModifier.Struct(context.Background(), s)
}
