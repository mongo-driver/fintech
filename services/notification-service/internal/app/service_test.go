package app

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/example/fintech-backend/shared/events"
)

type fakeSender struct {
	last events.NotificationEvent
}

func (f *fakeSender) Send(_ context.Context, event events.NotificationEvent) error {
	f.last = event
	return nil
}

type fakeConsumer struct {
	messages []kafka.Message
	index    int
	cancel   context.CancelFunc
}

func (f *fakeConsumer) Read(ctx context.Context) (kafka.Message, error) {
	if f.index >= len(f.messages) {
		return kafka.Message{}, context.Canceled
	}
	msg := f.messages[f.index]
	f.index++
	return msg, nil
}

func (f *fakeConsumer) Commit(context.Context, kafka.Message) error {
	if f.cancel != nil {
		f.cancel()
	}
	return nil
}

func TestSendManual(t *testing.T) {
	sender := &fakeSender{}
	svc := NewService(sender, zap.NewNop())
	err := svc.SendManual(context.Background(), "user-1", "Subject", "Message")
	require.NoError(t, err)
	require.Equal(t, "user-1", sender.last.UserID)
	require.Equal(t, "Subject", sender.last.Subject)
}

func TestHandleMessage(t *testing.T) {
	event := events.NewNotificationEvent("wallet_deposit", "u1", "Subject", "Message")
	payload, err := json.Marshal(event)
	require.NoError(t, err)

	sender := &fakeSender{}
	svc := NewService(sender, zap.NewNop())
	err = svc.handleMessage(context.Background(), kafka.Message{Value: payload})
	require.NoError(t, err)
	require.Equal(t, "u1", sender.last.UserID)
}

func TestConsume(t *testing.T) {
	event := events.NewNotificationEvent("wallet_deposit", "u2", "Subject", "Message")
	payload, err := json.Marshal(event)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sender := &fakeSender{}
	svc := NewService(sender, zap.NewNop())
	consumer := &fakeConsumer{
		messages: []kafka.Message{{Value: payload}},
		cancel:   cancel,
	}
	err = svc.Consume(ctx, consumer)
	require.NoError(t, err)
	require.Equal(t, "u2", sender.last.UserID)
}

func TestHandleMessageInvalidJSON(t *testing.T) {
	sender := &fakeSender{}
	svc := NewService(sender, zap.NewNop())
	err := svc.handleMessage(context.Background(), kafka.Message{Value: []byte("bad-json")})
	require.Error(t, err)
}
