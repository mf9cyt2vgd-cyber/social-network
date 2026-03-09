package usecase

import "notification-service/internal/domain"

type NotificationUsecase struct {
	Cache	domain.CacheRepository
	EventConsumer	domain.EventConsumer
}
