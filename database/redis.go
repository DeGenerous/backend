package database

import (
	"context"
	"github.com/go-redis/redis/v8"

	. "backend/config"
)

var (
	RedisContext = context.Background()
	RedisClient  *redis.Client
)

func redisInit() error {
	opt, err := redis.ParseURL(Config.Database.Redis)
	if err != nil {
		return err
	}

	RedisClient = redis.NewClient(opt)

	return nil
}
