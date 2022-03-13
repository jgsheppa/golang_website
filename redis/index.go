package redis

import (
	"context"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

type Client struct {
	client *redis.Client
}

func NewRedis() (*Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:        os.Getenv("REDIS_URL"),
		Password: "",
		DB:          0,
		DialTimeout: 100 * time.Millisecond,
		ReadTimeout: 100 * time.Millisecond,
	})

	if _, err := client.Ping(context.Background()).Result(); err != nil {
		return nil, err
	}

	return &Client{
		client: client,
	}, nil
}