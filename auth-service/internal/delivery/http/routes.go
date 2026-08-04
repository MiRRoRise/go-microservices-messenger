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
	r.Use(h.MetricsMiddleware)

	r.Get("/metrics", promhttp.Handler().ServeHTTP)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", h.Register)
		r.Post("/login", h.Login)

		r.Group(func(r chi.Router) {
			r.Use(h.AuthMiddleware)
			r.Get("/me", h.Me)
			r.Post("/logout", h.Logout)
		})
		r.Post("/refresh", h.Refresh)
	})
	r.Get("/swagger/*", httpSwagger.WrapHandler)
	
	return r
}
