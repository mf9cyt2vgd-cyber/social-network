package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"post-service/internal/domain"
	"time"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

type CacheRepository struct {
	*redis.Client
}

func NewRedisClient(addr string, db int, logger *slog.Logger) (*CacheRepository, error) {
	c := redis.NewClient(&redis.Options{
		Addr: addr, DB: db,
	})
	if err := redisotel.InstrumentTracing(c); err != nil {
		return nil, err
	}
	return &CacheRepository{
		Client: c,
	}, nil
}
func (r CacheRepository) List(ctx context.Context) ([]*domain.Post, error) {
	ids, err := r.SMembers(ctx, "posts:ids").Result()
	if err != nil {
		return nil, err
	}
	var posts []*domain.Post
	for _, id := range ids {
		post, err := r.GetPost(ctx, id)
		if err != nil {
			continue
		}
		posts = append(posts, post)
	}
	return posts, nil
}
func (r CacheRepository) GetPost(ctx context.Context, id string) (*domain.Post, error) {
	key := fmt.Sprintf("post:%s", id)

	var post domain.Post
	data, err := r.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get post by id %s: %w", id, err)
	}
	err = json.Unmarshal(data, &post)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal post from redis: %w", err)
	}
	return &post, err
}
func (r CacheRepository) SavePost(ctx context.Context, post *domain.Post, ttl time.Duration) error {
	key := fmt.Sprintf("post:%s", post.ID)
	postJSON, err := json.Marshal(post)
	if err != nil {
		return fmt.Errorf("failed to marshal post to redis: %w", err)
	}
	pipe := r.Pipeline()
	pipe.Set(ctx, key, postJSON, ttl)
	pipe.SAdd(ctx, "posts:ids", post.ID)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute pipeline: %w", err)
	}
	return nil
}
