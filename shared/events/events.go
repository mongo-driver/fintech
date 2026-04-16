package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type NotificationEvent struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"`
	UserID    string            `json:"user_id"`
	Channel   string            `json:"channel"`
	Subject   string            `json:"subject"`
	Message   string            `json:"message"`
	Metadata  map[string]string `json:"metadata"`
	CreatedAt time.Time         `json:"created_at"`
}

func NewNotificationEvent(eventType, userID, subject, message string) NotificationEvent {
	return NotificationEvent{
		ID:        uuid.NewString(),
		Type:      eventType,
		UserID:    userID,
		Channel:   "email",
		Subject:   subject,
		Message:   message,
		Metadata:  map[string]string{},
		CreatedAt: time.Now().UTC(),
	}
}

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string, topic string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(brokers...),
			Topic:                  topic,
			AllowAutoTopicCreation: true,
			Balancer:               &kafka.LeastBytes{},
			RequiredAcks:           kafka.RequireOne,
		},
	}
}

func (p *Producer) Publish(ctx context.Context, key string, payload any) error {
	if p == nil || p.writer == nil {
		return errors.New("producer not initialized")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: body,
		Time:  time.Now(),
	})
}

func (p *Producer) Close() error {
	if p == nil || p.writer == nil {
		return nil
	}
	return p.writer.Close()
}

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(brokers []string, topic, groupID string) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			GroupID:  groupID,
			Topic:    topic,
			MinBytes: 1,
			MaxBytes: 10e6,
		}),
	}
}

func (c *Consumer) Read(ctx context.Context) (kafka.Message, error) {
	if c == nil || c.reader == nil {
		return kafka.Message{}, errors.New("consumer not initialized")
	}
	return c.reader.FetchMessage(ctx)
}

func (c *Consumer) Commit(ctx context.Context, msg kafka.Message) error {
	return c.reader.CommitMessages(ctx, msg)
}

func (c *Consumer) Close() error {
	if c == nil || c.reader == nil {
		return nil
	}
	return c.reader.Close()
}
