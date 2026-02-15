package consumer

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

const (
	_defaultConnAttempts = 10
	_defaultConnTimeout  = time.Second
)

type Consumer struct {
	connAttempts int
	connTimeout  time.Duration

	brokers []string
	groupID string
	topic   string

	Reader *kafka.Reader
}

func New(ctx context.Context, brokers []string, groupID, topic string, opts ...Option) (*Consumer, error) {
	c := &Consumer{
		connAttempts: _defaultConnAttempts,
		connTimeout:  _defaultConnTimeout,
		brokers:      brokers,
		groupID:      groupID,
		topic:        topic,
	}

	for _, opt := range opts {
		opt(c)
	}

	c.Reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:  c.brokers,
		GroupID:  c.groupID,
		Topic:    c.topic,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	var err error

	for c.connAttempts > 0 {
		err = c.ping(ctx)
		if err == nil {
			break
		}

		log.Printf("Kafka consumer is trying to connect, attempts left: %d", c.connAttempts)

		time.Sleep(c.connTimeout)

		c.connAttempts--
	}

	if err != nil {
		return nil, fmt.Errorf("Kafka Consumer - New - connAttempts == 0: %w", err)
	}

	return c, nil
}

func (c *Consumer) ping(ctx context.Context) error {
	conn, err := kafka.DialContext(ctx, "tcp", c.brokers[0])
	if err != nil {
		return fmt.Errorf("Kafka Consumer - kafka.DialContext: %w", err)
	}
	defer conn.Close()

	_, err = conn.Brokers()
	if err != nil {
		return fmt.Errorf("Kafka Consumer - conn.Brokers: %w", err)
	}

	return nil
}

func (c *Consumer) Close() error {
	if c.Reader != nil {
		return c.Reader.Close()
	}
	return nil
}
