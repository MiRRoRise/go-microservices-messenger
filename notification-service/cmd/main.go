package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/MiRRoRise/notification-service/internal/kafka"
	"github.com/MiRRoRise/notification-service/internal/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	log := logger.New()

	brokersEnv := getEnv("KAFKA_BROKERS", "kafka:9092")
	brokers := strings.Split(brokersEnv, ",")
	httpAddr := getEnv("SERVER_PORT", ":8082")

	consumer, err := kafka.NewConsumer(brokers)
	if err != nil {
		log.Fatal("failed to create consumer", err)
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Error("failed to close consumer", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := consumer.Start(ctx, log); err != nil {
		log.Fatal("failed to start consumer", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{
		Addr:         httpAddr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Info("notification-service http started", "addr", httpAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("failed to start http server", err)
		}
	}()

	log.Info("notification-service started")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Info("shutting down...")
	cancel()

	select {
	case <-consumer.Done():
	case <-time.After(5 * time.Second):
		log.Info("consumer shutdown timed out")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("http server shutdown failed", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
