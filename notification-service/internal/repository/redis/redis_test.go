package redis

import (
	"context"
	"net/url"
	"notification-service/internal/domain"
	"strconv"
	"testing"

	"github.com/redis/go-redis/v9"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

func TestRedisCache_SaveNotificationWithLimit(t *testing.T) {
	ctx := context.Background()
	c, err := tcredis.Run(ctx, "redis:latest")
	if err != nil {
		t.Fatal("failed to start redis", err)
	}
	defer func() {
		if err := c.Terminate(ctx); err != nil {
			t.Fatal("failed to terminate container", err)
		}
	}()
	addr, err := c.ConnectionString(ctx)
	if err != nil {
		t.Fatal("failed to get connection string", err)
	}
	u, err := url.Parse(addr)
	if err != nil {
		t.Fatal("failed to parse redis connection string", err)
	}
	client := redis.NewClient(&redis.Options{Addr: u.Host})
	cache := &RedisCache{client}
	for i := 0; i < 110; i++ {
		post := &domain.Post{ID: strconv.Itoa(i), Content: "something"}
		err := cache.SaveNotificationWithLimit(ctx, post)
		if err != nil {
			t.Fatal("failed to save post to cache", err)
		}
	}
	count := client.ZCard(ctx, "notifications.set").Val()
	if count != 100 {
		t.Fatal("expected 100, got", count)
	}
}
