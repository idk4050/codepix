package pixkey

type Type uint8

const (
	CPFKey Type = iota + 1
	PhoneKey
	EmailKey
)

type Key = string

type PixKey struct {
	Type Type
	Key  Key
}
