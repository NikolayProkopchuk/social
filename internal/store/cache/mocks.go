package cache

import (
	"context"

	"github.com/NikolayProkopchuk/social/internal/store"
	"github.com/stretchr/testify/mock"
)

func NewMockCache() *Cache {
	return &Cache{
		Users: &MockUserCache{},
	}
}

type MockUserCache struct {
	mock.Mock
}

func (m *MockUserCache) Get(ctx context.Context, id int64) (*store.User, error) {
	args := m.Called(ctx, id)
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	return result.(*store.User), args.Error(1)
}

func (m *MockUserCache) Set(ctx context.Context, u store.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}
