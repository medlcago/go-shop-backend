package token

import "errors"

type ErrTokenError struct {
	Err error `json:"error"`
}

func (e ErrTokenError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}

	return "invalid token"
}

func IsErrTokenError(err error) bool {
	_, ok := errors.AsType[*ErrTokenError](err)
	return ok
}
