package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"post-service/internal/domain"
	"time"

	"github.com/google/uuid"
)

type PostUsecase struct {
	repo  domain.PostRepository  // db - postgres
	cache domain.CacheRepository // redis
}

func NewPostUsecase(poolRepo domain.PostRepository, cache domain.CacheRepository) *PostUsecase {
	return &PostUsecase{
		repo:  poolRepo,
		cache: cache,
	}
}

func (u *PostUsecase) List(ctx context.Context) ([]*domain.Post, error) {
	posts, _ := u.cache.List(ctx)
	if posts != nil {
		return posts, nil
	}
	posts, err := u.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list posts: %w", err)
	}

	return posts, nil
}

func (u *PostUsecase) GetByID(ctx context.Context, id string) (*domain.Post, error) {
	post, _ := u.cache.GetPost(ctx, id)
	if post != nil {
		return post, nil
	}

	post, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get post by id %s: %w", id, err)
	}

	go func() {
		cacheCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		_ = u.cache.SavePost(cacheCtx, post, 5*time.Minute)
	}()

	return post, nil
}

func (u *PostUsecase) CreatePost(ctx context.Context, title, author, content string, tags []string) (*domain.Post, error) {
	post := &domain.Post{
		ID:        uuid.NewString(),
		Title:     title,
		Author:    author,
		Content:   content,
		Tags:      tags,
		CreatedAt: time.Now(),
	}

	if err := u.repo.Save(ctx, post); err != nil {
		return nil, fmt.Errorf("failed to save post to database: %w", err)
	}

	go func(p *domain.Post) {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		err := u.cache.SavePost(cacheCtx, post, 5*time.Minute)
		if err != nil {
			slog.Error("failed to save in cache", "error", err)
		}
	}(post)

	return post, nil
}
