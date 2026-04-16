package app

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"github.com/example/fintech-backend/shared/events"
)

type Sender interface {
	Send(ctx context.Context, event events.NotificationEvent) error
}

type LogSender struct {
	log *zap.Logger
}

func NewLogSender(log *zap.Logger) *LogSender {
	return &LogSender{log: log}
}

func (s *LogSender) Send(_ context.Context, event events.NotificationEvent) error {
	s.log.Info("notification dispatched",
		zap.String("event_id", event.ID),
		zap.String("user_id", event.UserID),
		zap.String("type", event.Type),
		zap.String("channel", event.Channel),
		zap.String("subject", event.Subject),
		zap.String("message", event.Message),
	)
	return nil
}

type Service struct {
	sender Sender
	log    *zap.Logger
}

func NewService(sender Sender, log *zap.Logger) *Service {
	return &Service{sender: sender, log: log}
}

type Consumer interface {
	Read(ctx context.Context) (kafka.Message, error)
	Commit(ctx context.Context, msg kafka.Message) error
}

func (s *Service) Consume(ctx context.Context, consumer Consumer) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		msg, err := consumer.Read(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			s.log.Warn("failed to read kafka message", zap.Error(err))
			continue
		}

		if err = s.handleMessage(ctx, msg); err != nil {
			s.log.Error("failed to send notification", zap.Error(err))
			continue
		}
		if err = consumer.Commit(ctx, msg); err != nil {
			s.log.Warn("failed to commit kafka message", zap.Error(err))
		}
	}
}

func (s *Service) handleMessage(ctx context.Context, msg kafka.Message) error {
	var event events.NotificationEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return err
	}
	return s.sender.Send(ctx, event)
}

func (s *Service) SendManual(ctx context.Context, userID, subject, message string) error {
	event := events.NewNotificationEvent("manual_notification", userID, subject, message)
	if event.Channel == "" {
		return fmt.Errorf("invalid channel")
	}
	return s.sender.Send(ctx, event)
}
