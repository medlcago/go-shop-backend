package hasher

import "github.com/stretchr/testify/mock"

var _ Hasher = (*MockHasher)(nil)

type MockHasher struct {
	mock.Mock
}

func (h *MockHasher) Hash(password string) (string, error) {
	args := h.Called(password)
	return args.String(0), args.Error(1)
}

func (h *MockHasher) Verify(password string, hash string) (bool, error) {
	args := h.Called(password, hash)
	return args.Bool(0), args.Error(1)
}
