package token

import "github.com/stretchr/testify/mock"

var _ Manager = (*ManagerMock)(nil)

type ManagerMock struct {
	mock.Mock
}

func (t *ManagerMock) GenerateAccessToken(payload Payload) (string, error) {
	args := t.Called(payload)
	return args.String(0), args.Error(1)
}

func (t *ManagerMock) GenerateRefreshToken(payload Payload) (string, error) {
	args := t.Called(payload)
	return args.String(0), args.Error(1)
}

func (t *ManagerMock) ValidateToken(tokenString string) (*UserClaims, error) {
	args := t.Called(tokenString)
	return args.Get(0).(*UserClaims), args.Error(1)
}
