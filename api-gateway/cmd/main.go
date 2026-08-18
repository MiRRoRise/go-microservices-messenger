package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	idSegment = regexp.MustCompile(`/\d+`)

	requestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"service", "method", "endpoint", "status"},
	)
	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"service", "method", "endpoint"},
	)
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	authURL := getEnv("AUTH_SERVICE_URL", "http://auth-service:8080")
	chatURL := getEnv("CHAT_SERVICE_URL", "http://chat-service:8081")
	addr := getEnv("SERVER_PORT", ":8000")

	authProxy, err := newReverseProxy(authURL)
	if err != nil {
		logger.Error("invalid auth service url", "err", err, "url", authURL)
		os.Exit(1)
	}
	chatProxy, err := newReverseProxy(chatURL)
	if err != nil {
		logger.Error("invalid chat service url", "err", err, "url", chatURL)
		os.Exit(1)
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"status":"ok"}`)
	})
	r.Handle("/metrics", promhttp.Handler())

	r.Group(func(r chi.Router) {
		r.Use(metricsMiddleware)

		r.Handle("/auth/*", authProxy)
		r.Handle("/auth", authProxy)
		r.Handle("/chats/*", chatProxy)
		r.Handle("/chats", chatProxy)
	})

	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("api-gateway started", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "err", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Info("shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server shutdown failed", "err", err)
	}
	logger.Info("api-gateway stopped")
}

func newReverseProxy(rawURL string) (http.Handler, error) {
	target, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	original := proxy.Director
	proxy.Director = func(req *http.Request) {
		original(req)
		req.Host = target.Host
	}
	return proxy, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)

		service := upstreamService(path)
		endpoint := normalizePath(path)
		status := strconv.Itoa(sw.status)

		requestsTotal.WithLabelValues(service, r.Method, endpoint, status).Inc()
		requestDuration.WithLabelValues(service, r.Method, endpoint).Observe(time.Since(start).Seconds())
	})
}

func upstreamService(path string) string {
	switch {
	case strings.HasPrefix(path, "/auth"):
		return "auth"
	case strings.HasPrefix(path, "/chats"):
		return "chat"
	default:
		return "gateway"
	}
}

func normalizePath(path string) string {
	return idSegment.ReplaceAllString(path, "/{id}")
}
