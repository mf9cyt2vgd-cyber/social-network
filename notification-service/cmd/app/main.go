package main

import (
	"context"
	"notification-service/internal/config"
	"notification-service/internal/logger"
	"notification-service/internal/repository/redis"
	"notification-service/internal/transport/kafka"
	"notification-service/internal/usecase"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.MustLoad()
	log := logger.SetupLogger(cfg.Env)
	ctxC, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumer, err := kafka.NewKafkaConsumer(cfg.Brokers, cfg.GroupID, cfg.Topic, log)
	if err != nil {
		log.Error("failed to create Kafka consumer")
		os.Exit(1)
	}
	defer func(consumer *kafka.KafkaConsumer) {
		err := consumer.Close()
		if err != nil {
			log.Error("error closing Kafka consumer", "error", err)
		}
	}(consumer)

	cache := redis.New(cfg.Redis.Addr, cfg.Redis.DB, log)
	defer cache.Close()

	notificationsUC := usecase.NotificationUsecase{
		Cache:         cache,
		EventConsumer: consumer,
	}
	go func() {
		notificationsUC.StartSendingNotifications(ctxC, log)
	}()
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-done
	log.Info("got interrupt signal, stopping...")
	cancel()
}
