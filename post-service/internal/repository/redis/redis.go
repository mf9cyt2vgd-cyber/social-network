package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"post-service/internal/domain"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheRepository struct {
	*redis.Client
}

func NewRedisClient(addr string, db int, logger *slog.Logger) (*CacheRepository, error) {
	c := redis.NewClient(&redis.Options{
		Addr: addr, DB: db,
	})
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
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, nil
}
func (r CacheRepository) GetPost(ctx context.Context, id string) (*domain.Post, error) {
	key := fmt.Sprintf("post:%s", id)

	var post domain.Post
	data, _ := r.Get(ctx, key).Bytes()
	err := json.Unmarshal(data, &post)
	if err != nil {
		return nil, fmt.Errorf("failed to get post by id %s: %w", id, err)
	}
	return &post, err
}
func (r CacheRepository) SavePost(ctx context.Context, post *domain.Post, ttl time.Duration) error {
	key := fmt.Sprintf("post:%s", post.ID)
	postJSON, err := json.Marshal(post)
	if err != nil {
		return err
	}
	err = r.Set(ctx, key, postJSON, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to save post to redis: %w", err)
	}
	err = r.SAdd(ctx, "posts:ids", post.ID).Err()
	if err != nil {
		return fmt.Errorf("failed to save post.id to set in redis: %w", err)
	}
	return nil
}
