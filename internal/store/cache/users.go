package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/NikolayProkopchuk/social/internal/store"
	"github.com/go-redis/redis/v8"
)

type userCache struct {
	client *redis.Client
}

const userExpiration = time.Hour * 3

func (cache *userCache) Get(ctx context.Context, id int64) (*store.User, error) {
	key := fmt.Sprintf("user-%d", id)
	data, err := cache.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, store.ErrNotFound
		}
		return nil, err
	}
	var user store.User
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (cache *userCache) Set(ctx context.Context, user store.User) error {
	key := fmt.Sprintf("user-%d", user.ID)
	// Serialize user to JSON
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	status := cache.client.Set(ctx, key, data, userExpiration)
	return status.Err()
}
