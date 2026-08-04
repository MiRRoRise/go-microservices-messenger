package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/MiRRoRise/notification-service/internal/logger"
)

const (
	TopicMessageCreated = "message.created"
	TopicChatCreated    = "chat.created"
	TopicUserRegistered = "user.registered"
)

type MessageCreatedEvent struct {
	MessageID int64     `json:"message_id"`
	ChatID    int64     `json:"chat_id"`
	SenderID  int64     `json:"sender_id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

type ChatCreatedEvent struct {
	ChatID    int64     `json:"chat_id"`
	Name      string    `json:"name"`
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
	topics := []string{TopicMessageCreated, TopicChatCreated, TopicUserRegistered}

	var wg sync.WaitGroup
	for _, topic := range topics {
		partitionConsumer, err := c.consumer.ConsumePartition(topic, 0, sarama.OffsetOldest)
		if err != nil {
			return fmt.Errorf("failed to start partition consumer: %w", err)
		}
		wg.Add(1)

		go func(topic string, pc sarama.PartitionConsumer) {
			defer wg.Done()
			defer pc.Close()

			for {
				select {
				case msg, ok := <-pc.Messages():
					if !ok {
						return 
					}
					c.handleMessage(topic, msg, logger)
				case err := <-pc.Errors():
					logger.Error("failed partition consumer", err)
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

func (c *Consumer) Close() error {
	return c.consumer.Close()
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
		
	case TopicChatCreated:
		var event ChatCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			logger.Error("failed to unmarshal event", err)
			return
		}

		logger.Info(
			"New chat",
			"chat_id", event.ChatID,
			"name", event.Name,
		)

	case TopicUserRegistered:
		var event UserRegisteredEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			logger.Error("failed to unmarshal event", err)
			return
		}

		logger.Info(
			"New user",
			"user_id", event.UserID,
			"email", event.Email,
		)

	default: 
		logger.Info("unknown topic")
	}
}
