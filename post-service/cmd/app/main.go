// cmd/main.go
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"post-service/internal/repository/events/forwarder"
	"post-service/internal/repository/redis"
	"syscall"
	"time"

	"post-service/internal/repository/postgres"
	route "post-service/internal/router"

	"post-service/internal/config"
	"post-service/internal/lib/logger"
	"post-service/internal/usecase"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	// Configurate system
	cfg := config.MustLoad()

	// Settings logger
	log := logger.SetupLogger(cfg.Env)
	log.Info("starting the project...", slog.String("env", cfg.Env))

	// Подключаемся к БД
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Error("failed to connect to db:", slog.Any("err", err))
	}
	defer pool.Close()

	postRepo, err := postgres.NewPostgresPostRepository(pool, log)
	if err != nil {
		log.Error("failed to create post repository", "error", err)
		os.Exit(1)
	}
	defer postRepo.Close()

	cache, err := redis.NewRedisClient(cfg.Redis.Addr, cfg.Redis.DB, log.With(slog.String("component", "redis")))
	defer cache.Close()

	postUC := usecase.NewPostUsecase(postRepo, cache) // Бизнес-логика для posts

	fwd, err := forwarder.StartForwarder(ctx, pool, forwarder.Config{KafkaBrokers: cfg.Brokers, Logger: log})
	if err != nil {
		log.Error("failed to start outbox forwarder", "error", err)
		return
	}
	// Передаем ctx в обработчики
	router := route.New(ctx, log.With(slog.String("component", "http")), postUC)

	// Settings and started server + Graceful shutdown
	srv := &http.Server{
		Addr:         cfg.Address,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
		Handler:      router,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	log.Info("server starting", slog.String("address", cfg.Address))

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server failed to start", slog.Any("error", err))
			os.Exit(1)
		}
		log.Info("server stopped listening") // Когда выйдет из ListenAndServe
	}()

	<-done
	log.Info("server stopping...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("server forced to shutdown", slog.Any("error", err))
		os.Exit(1)
	}

	if err := fwd.Stop(); err != nil {
		log.Error("failed to stop forwarder", "error", err)
	}

	log.Info("server stopped gracefully")
}
