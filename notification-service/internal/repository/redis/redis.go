package redis

import "github.com/redis/go-redis/v9"

type RedisCache struct {
	client       *redis.Client
	maxSize      int
	sortedSetKey string
}
