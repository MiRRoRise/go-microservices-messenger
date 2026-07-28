package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/MiRRoRise/notification-service/internal/kafka"
	"github.com/MiRRoRise/notification-service/internal/logger"
)

func main() {
	logger := logger.New()

	consumer, err := kafka.NewConsumer([]string{"kafka:9092"})
	if err != nil {
		logger.Fatal("failed to create consumer", err)
	}
	defer consumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := consumer.Start(ctx, logger); err != nil {
		logger.Fatal("failed to start consumer", err)
	}

	logger.Info("notification-service started")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Info("shutting down...")
	cancel()
}
