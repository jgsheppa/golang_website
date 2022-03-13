package redis

import (
	"context"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
)

type Client struct {
	client *redis.Client
}

func NewRedis() (*Client, error) {
	redisURL := os.Getenv("REDIS_URL")
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatal(err)
	}

	client := redis.NewClient(opt)

	if _, err := client.Ping(context.Background()).Result(); err != nil {
		return nil, err
	}

	return &Client{
		client: client,
	}, nil
}