package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/MiRRoRise/notification-service/internal/logger"
)

type MessageCreatedEvent struct {
	MessageID int64     `json:"message_id"`
	ChatID    int64     `json:"chat_id"`
	SenderID  int64     `json:"sender_id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

type Consumer struct {
	consumer sarama.Consumer
	done     chan struct{}
}

func NewConsumer(brokers []string) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	return &Consumer{
		consumer: consumer,
		done:     make(chan struct{}),
	}, nil
}

func (c *Consumer) Start(ctx context.Context, logger *logger.Logger) error {
	partitionConsumer, err := c.consumer.ConsumePartition("message.created", 0, sarama.OffsetOldest)
	if err != nil {
		return fmt.Errorf("failed to start partition consumer: %w", err)
	}

	go func() {
		for {
			select {
			case msg := <-partitionConsumer.Messages():
				c.handleMessage(msg, logger)
			case err := <-partitionConsumer.Errors():
				logger.Error("failed partition consumer", err)
			case <-ctx.Done():
				partitionConsumer.Close()
				close(c.done)
				return
			}
		}
	}()

	return nil
}

func (c *Consumer) Close() error {
	return c.consumer.Close()
}

func (c *Consumer) handleMessage(msg *sarama.ConsumerMessage, logger *logger.Logger) {
	var event MessageCreatedEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		logger.Error("failed to unmarshal event", err)
		return
	}

	logger.Info(
		"New message received",
		"message_id", event.MessageID,
		"chat_id", event.ChatID,
		"sender_id", event.SenderID,
		"text", event.Text,
	)
}
