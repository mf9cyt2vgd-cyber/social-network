package redis

import (
	"context"
	"log/slog"
	"notification-service/internal/domain"
	"notification-service/internal/mapper"
	"time"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func New(addr string, db int, log *slog.Logger) *RedisCache {
	c := redis.NewClient(&redis.Options{Addr: addr, DB: db})
	if err := redisotel.InstrumentTracing(c); err != nil {
		return nil
	}
	return &RedisCache{
		client: c,
	}
}
func (r *RedisCache) SaveNotificationWithLimit(ctx context.Context, post *domain.Post) error {
	pipe := r.client.Pipeline()
	score := float64(time.Now().UnixMilli())
	pipe.ZAdd(
		ctx,
		"notifications.set",
		redis.Z{
			Score:  score,
			Member: mapper.ConvertPostToNotification(post)})
	pipe.ZRemRangeByRank(ctx, "notifications.set", 0, -101)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}
func (r *RedisCache) Close() error {
	err := r.client.Close()
	if err != nil {
		return err
	}
	return nil
}
