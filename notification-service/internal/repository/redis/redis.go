package redis

import (
	"context"
	"log/slog"
	"notification-service/internal/domain"
	"notification-service/internal/mapper"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func New(addr string, db int, log *slog.Logger) *RedisCache {
	c := redis.NewClient(&redis.Options{Addr: addr, DB: db})
	return &RedisCache{
		client: c,
	}
}
func (r *RedisCache) SaveNotificationWithLimit(ctx context.Context, post *domain.Post) error {
	pipe := r.client.Pipeline()
	score := float64(time.Now().UnixMilli())
	err := pipe.ZAdd(
		ctx,
		"notifications.set",
		redis.Z{
			Score:  score,
			Member: mapper.ConvertPostToNotification(post)}).Err()
	if err != nil {
		return err
	}
	err = pipe.ZRemRangeByRank(ctx, "notifications.set", 0, -101).Err()
	if err != nil {
		return err
	}
	_, err = pipe.Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}
func (r *RedisCache) Close() {
	r.client.Close()
}
