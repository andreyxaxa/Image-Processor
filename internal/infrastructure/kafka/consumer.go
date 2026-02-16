package kafka

import (
	"context"
	"fmt"

	"github.com/andreyxaxa/Image-Processor/pkg/kafka/consumer"
	"github.com/segmentio/kafka-go"
)

type EventConsumer struct {
	*consumer.Consumer
}

func NewEventConsumer(consumer *consumer.Consumer) *EventConsumer {
	return &EventConsumer{consumer}
}

func (ec *EventConsumer) ReadEvent(ctx context.Context) (kafka.Message, error) {
	msg, err := ec.Reader.FetchMessage(ctx)
	if err != nil {
		return kafka.Message{}, fmt.Errorf("EventConsumer - ReadEvent - ec.Reader.FetchMessage: %w", err)
	}

	return msg, nil
}

func (ec *EventConsumer) CommitEvent(ctx context.Context, event kafka.Message) error {
	err := ec.Reader.CommitMessages(ctx, event)
	if err != nil {
		return fmt.Errorf("EventConsumer - CommitEvent - ec.Reader.CommitMessages: %w", err)
	}

	return nil
}

func (ec *EventConsumer) Close() error {
	err := ec.Consumer.Close()
	if err != nil {
		return fmt.Errorf("EventConsumer - Close: %w", err)
	}

	return nil
}
