package kafka

import (
	"context"
	"fmt"

	"github.com/andreyxaxa/Image-Processor/internal/entity"
	"github.com/andreyxaxa/Image-Processor/pkg/kafka/producer"
	"github.com/segmentio/kafka-go"
)

type EventProducer struct {
	*producer.Producer
	maxRetries int
	topic      string
}

func NewEventProducer(producer *producer.Producer, retries int, topic string) *EventProducer {
	return &EventProducer{
		producer,
		retries,
		topic,
	}
}

func (ep *EventProducer) SendEvents(ctx context.Context, events []*entity.OutboxEvent) error {
	var msgsToSend []kafka.Message

	for _, event := range events {
		msg := kafka.Message{
			Topic: ep.topic,
			Key:   []byte(event.AggregateID.String()),
			Value: event.Payload,
			Headers: []kafka.Header{
				{Key: "event_id", Value: []byte(event.ID.String())},
			},
		}
		msgsToSend = append(msgsToSend, msg)
	}

	if len(msgsToSend) == 0 {
		return nil
	}

	err := ep.Writer.WriteMessages(ctx, msgsToSend...)
	if err != nil {
		return fmt.Errorf("EventProducer - SendEvents - ep.Writer.WriteMessages: %w", err)
	}

	return nil
}

func (ep *EventProducer) Close() error {
	err := ep.Producer.Close()
	if err != nil {
		return fmt.Errorf("EventProducer - Close: %w", err)
	}

	return nil
}
