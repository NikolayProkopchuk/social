package store

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockUserStore struct {
	mock.Mock
}

func (m *MockUserStore) Activate(ctx context.Context, code string) error {
	panic("unimplemented")
}

func (m *MockUserStore) GetByID(ctx context.Context, id int64) (*User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	panic("unimplemented")
}

func (m *MockUserStore) CreateAndInvite(ctx context.Context, user *User, baseURL string, activationTTL time.Duration) error {
	panic("unimplemented")
}
