package router

import (
	"context"
	"log/slog"

	httpDelivery "post-service/internal/delivery/http"
	"post-service/internal/usecase"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/riandyrn/otelchi"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// New инициализирует chi.Router
func New(
	ctx context.Context,
	log *slog.Logger,
	postUC *usecase.PostUsecase,
	// commentUC *usecase.CommentUsecase,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(otelchi.Middleware("post-service", otelchi.WithChiRoutes(r)))

	// Middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)

	// Handlers
	r.Handle("/metrics", promhttp.Handler())
	postHandler := httpDelivery.NewPostHandler(log, postUC)
	healthHandler := httpDelivery.NewHealthHandler(log)
	// commentHandler := httpDelivery.NewCommentHandler(log, commentUC)

	// Routes
	r.Route("/posts", func(r chi.Router) {
		r.Post("/", postHandler.CreatePost)
		r.Get("/", postHandler.List)
		r.Get("/{id}", postHandler.GetByID)
	})
	r.Get("/health", healthHandler.HealthCheck)

	// r.Route("/comments", func(r chi.Router) {
	// 	r.Post("/", commentHandler.CreateComment)
	// })

	return r
}
