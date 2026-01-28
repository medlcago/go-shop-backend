package token

import "github.com/stretchr/testify/mock"

var _ Manager = (*ManagerMock)(nil)

type ManagerMock struct {
	mock.Mock
}

func (t *ManagerMock) GenerateAccessToken(payload map[string]any) (string, error) {
	args := t.Called(payload)
	return args.String(0), args.Error(1)
}

func (t *ManagerMock) GenerateRefreshToken(payload map[string]any) (string, error) {
	args := t.Called(payload)
	return args.String(0), args.Error(1)
}

func (t *ManagerMock) ValidateToken(tokenString string) (map[string]interface{}, error) {
	args := t.Called(tokenString)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}
