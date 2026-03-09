package domain

import "context"

type CacheRepository interface {
	SaveNotificationWithLimit(ctx context.Context, post *Post) error
}
