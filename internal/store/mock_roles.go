package store

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockRoleStore struct {
	mock.Mock
}

func (s *MockRoleStore) GetByName(ctx context.Context, name string) (*Role, error) {
	args := s.Called(ctx, name)
	return args.Get(0).(*Role), args.Error(1)
}
