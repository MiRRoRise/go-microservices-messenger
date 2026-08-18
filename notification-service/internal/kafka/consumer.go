package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/MiRRoRise/notification-service/internal/logger"
	"github.com/MiRRoRise/notification-service/internal/metrics"
)

const (
	TopicMessageCreated = "message.created"
	TopicUserRegistered = "user.registered"
)

type MessageCreatedEvent struct {
	MessageID int64     `json:"message_id"`
	ChatID    int64     `json:"chat_id"`
	SenderID  int64     `json:"sender_id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

type UserRegisteredEvent struct {
	UserID    int64     `json:"user_id"`
	Email     string    `json:"email"`
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
	topics := []string{TopicMessageCreated, TopicUserRegistered}

	var wg sync.WaitGroup
	for _, topic := range topics {
		partitionConsumer, err := c.consumePartitionWithRetry(ctx, topic)
		if err != nil {
			return fmt.Errorf("failed to start partition consumer for %s: %w", topic, err)
		}
		wg.Add(1)

		go func(topic string, pc sarama.PartitionConsumer) {
			defer wg.Done()
			defer func() {
				if err := pc.Close(); err != nil {
					logger.Error("failed to close partition consumer", err)
				}
			}()

			for {
				select {
				case msg, ok := <-pc.Messages():
					if !ok {
						return
					}
					c.handleMessage(topic, msg, logger)
				case err := <-pc.Errors():
					if err != nil {
						logger.Error("failed partition consumer", err)
					}
				case <-ctx.Done():
					return
				}
			}
		}(topic, partitionConsumer)
	}

	go func() {
		wg.Wait()
		close(c.done)
	}()

	return nil
}

func (c *Consumer) consumePartitionWithRetry(ctx context.Context, topic string) (sarama.PartitionConsumer, error) {
	var lastErr error
	for attempt := 0; attempt < 30; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		pc, err := c.consumer.ConsumePartition(topic, 0, sarama.OffsetOldest)
		if err == nil {
			return pc, nil
		}
		lastErr = err
		time.Sleep(2 * time.Second)
	}
	return nil, lastErr
}

func (c *Consumer) Close() error {
	return c.consumer.Close()
}

func (c *Consumer) Done() <-chan struct{} {
	return c.done
}

func (c *Consumer) handleMessage(topic string, msg *sarama.ConsumerMessage, logger *logger.Logger) {
	switch topic {
	case TopicMessageCreated:
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
		metrics.KafkaEventsConsumed.WithLabelValues(topic).Inc()

	case TopicUserRegistered:
		var event UserRegisteredEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			logger.Error("failed to unmarshal event", err)
			return
		}

		logger.Info(
			"User registered",
			"user_id", event.UserID,
			"email", event.Email,
		)
		metrics.KafkaEventsConsumed.WithLabelValues(topic).Inc()

	default:
		logger.Info("unknown topic", "topic", topic)
	}
}
