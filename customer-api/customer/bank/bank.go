package bank

import "codepix/customer-api/customer"

type Code = uint32

type Bank struct {
	Code     Code
	Name     string
	Customer customer.Customer
}

func New(code Code, name string, customer customer.Customer) (*Bank, error) {
	return &Bank{code, name, customer}, nil
}
