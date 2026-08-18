package kafka

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
)

const TopicUserRegistered = "user.registered"

type UserRegisteredEvent struct {
	UserID    int64     `json:"user_id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type EventPublisher interface {
	PublishUserRegistered(event UserRegisteredEvent) error
}

type Producer struct {
	producer sarama.SyncProducer
}

func NewProducer(brokers []string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	return &Producer{producer: producer}, nil
}

func (p *Producer) Close() error {
	return p.producer.Close()
}

func (p *Producer) PublishUserRegistered(event UserRegisteredEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: TopicUserRegistered,
		Value: sarama.ByteEncoder(data),
	}

	_, _, err = p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

type NoopPublisher struct{}

func (NoopPublisher) PublishUserRegistered(UserRegisteredEvent) error {
	return nil
}
