package usecase

import (
	"context"
	"log/slog"
	"notification-service/internal/domain"

	"go.opentelemetry.io/otel/trace"
)

type NotificationUsecase struct {
	Cache         domain.CacheRepository
	EventConsumer domain.EventConsumer
}

func (n *NotificationUsecase) StartSendingNotifications(ctx context.Context, log *slog.Logger) {
	posts := n.EventConsumer.Consume(ctx)
	for postEvent := range posts {
		n.handleSinglePost(postEvent, log)
	}
}
func (n *NotificationUsecase) handleSinglePost(event *domain.PostEvent, log *slog.Logger) {
	incomingCtx := event.TraceCtx
	span := trace.SpanFromContext(incomingCtx)
	defer span.End()

	safeCtx := context.WithoutCancel(incomingCtx)

	log.Info("received post from Kafka", "post_id", event.Post.ID)

	err := n.Cache.SaveNotificationWithLimit(safeCtx, event.Post)

	if err != nil {
		span.RecordError(err)

		log.Error("failed to save notification in cache", "error", err)
		return
	}

}
