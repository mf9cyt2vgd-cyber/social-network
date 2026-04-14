package main

import (
	"context"
	"fmt"
	"notification-service/internal/config"
	"notification-service/internal/lib/gotel"
	"notification-service/internal/lib/logger"
	"notification-service/internal/repository/redis"
	"notification-service/internal/transport/kafka"
	"notification-service/internal/usecase"
	"os"
	"os/signal"
	"syscall"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func init() {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, // Важно для парсинга traceparent
		propagation.Baggage{},
	))
}
func main() {
	cfg := config.MustLoad()
	log := logger.SetupLogger(cfg.Env)
	ctxC, cancel := context.WithCancel(context.Background())
	defer cancel()
	shutdown := gotel.InitTracer(ctxC)
	defer func(ctx context.Context) {
		err := shutdown(ctxC)
		if err != nil {
			log.Error("failed to stop tracer", "error", err)
		}
	}(ctxC)
	consumer, err := kafka.NewKafkaConsumer(cfg.Kafka.Brokers, cfg.Kafka.GroupID, cfg.Kafka.Topic, log)
	fmt.Println(cfg.Kafka.GroupID)
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
	defer func(cache *redis.RedisCache) {
		err = cache.Close()
		if err != nil {
			log.Error("failed to close cache", "error", err)
		}
	}(cache)

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
