package cache

import (
	"github.com/go-redis/redis/v8"
)

func NewRedisClient(addr, password string, db int) *redis.Client {
	opts := &redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	}
	return redis.NewClient(opts)
}

func NewRedisStorage(client *redis.Client) *Cache {
	return &Cache{
		Users: &userCache{client: client},
	}
}
