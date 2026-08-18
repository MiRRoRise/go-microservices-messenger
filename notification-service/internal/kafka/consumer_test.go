package kafka

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/MiRRoRise/notification-service/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleMessage_MessageCreated(t *testing.T) {
	c := &Consumer{}
	log := logger.New()

	event := MessageCreatedEvent{
		MessageID: 1,
		ChatID:    2,
		SenderID:  3,
		Text:      "hello",
		CreatedAt: time.Now().UTC(),
	}
	raw, err := json.Marshal(event)
	require.NoError(t, err)

	assert.NotPanics(t, func() {
		c.handleMessage(TopicMessageCreated, &sarama.ConsumerMessage{Value: raw}, log)
	})
}

func TestHandleMessage_UserRegistered(t *testing.T) {
	c := &Consumer{}
	log := logger.New()

	event := UserRegisteredEvent{
		UserID:    10,
		Email:     "a@b.com",
		CreatedAt: time.Now().UTC(),
	}
	raw, err := json.Marshal(event)
	require.NoError(t, err)

	assert.NotPanics(t, func() {
		c.handleMessage(TopicUserRegistered, &sarama.ConsumerMessage{Value: raw}, log)
	})
}

func TestHandleMessage_InvalidJSON(t *testing.T) {
	c := &Consumer{}
	log := logger.New()

	assert.NotPanics(t, func() {
		c.handleMessage(TopicMessageCreated, &sarama.ConsumerMessage{Value: []byte("not-json")}, log)
	})
}
