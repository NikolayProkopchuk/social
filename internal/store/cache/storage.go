package cache

import (
	"context"

	"github.com/NikolayProkopchuk/social/internal/store"
)

type Storage struct {
	Users interface {
		Get(ctx context.Context, id int64) (*store.User, error)
		Set(ctx context.Context, u store.User) error
	}
}
