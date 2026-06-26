package token

import "errors"

type ErrInvalidToken struct {
	Err error `json:"error"`
}

func (e ErrInvalidToken) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}

	return "invalid token"
}

func IsErrInvalidToken(err error) bool {
	_, ok := errors.AsType[*ErrInvalidToken](err)
	return ok
}
