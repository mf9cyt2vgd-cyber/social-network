package usecase

import (
	"context"
	"log/slog"
	"notification-service/internal/domain"
)

type NotificationUsecase struct {
	Cache         domain.CacheRepository
	EventConsumer domain.EventConsumer
}

func (n *NotificationUsecase) StartSendingNotifications(ctx context.Context, log *slog.Logger) {
	posts := n.EventConsumer.Consume(ctx)
	for post := range posts {
		log.Info("received post from Kafka", "post_id", post.ID)
		err := n.Cache.SaveNotificationWithLimit(ctx, post)
		if err != nil {
			log.Error("failed to save notification in cache", "error", err)
			continue
		}
	}
}
