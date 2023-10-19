package helpers

import (
	"net/mail"
)

func GetValidEmail(email string) (string, error) {
	e, err := mail.ParseAddress(email)
	if err != nil {
		return "", err
	}
	return e.Address, nil
}

