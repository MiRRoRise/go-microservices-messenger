package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
)

func (h *Handler) RegisterRoutes() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	r.Get("/health", h.HealthCheck)
	r.Get("/metrics", promhttp.Handler().ServeHTTP)

	r.Group(func(r chi.Router) {
		r.Use(h.MetricsMiddleware)

		r.Get("/swagger/*", httpSwagger.WrapHandler)

		r.Route("/chats", func(r chi.Router) {
			r.Use(h.AuthMiddleware)

			r.Post("/", h.CreateChat)
			r.Get("/", h.ListChats)
			r.Get("/{id}", h.GetChatByID)

			r.Route("/{id}/messages", func(r chi.Router) {
				r.Post("/", h.CreateMessage)
				r.Get("/", h.ListMessages)
			})
		})
	})

	return r
}
