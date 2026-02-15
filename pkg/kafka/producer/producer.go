package producer

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

type Producer struct {
	connAttempts int
	connTimeout  time.Duration

	brokers []string
	Writer  *kafka.Writer
}

func New(ctx context.Context, brokers []string, opts ...Option) (*Producer, error) {
	p := &Producer{
		connAttempts: _defaultConnAttempts,
		connTimeout:  _defaultConnTimeout,
		brokers:      brokers,
	}

	for _, opt := range opts {
		opt(p)
	}

	p.Writer = &kafka.Writer{
		Addr:         kafka.TCP(p.brokers...),
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
	}

	var err error
	for p.connAttempts > 0 {
		err = p.ping(ctx)
		if err == nil {
			break
		}

		log.Printf("Kafka producer is trying to connect, attempts left: %d", p.connAttempts)

		time.Sleep(p.connTimeout)

		p.connAttempts--
	}

	if err != nil {
		return nil, fmt.Errorf("Kafka Producer - New - connAttempts == 0: %w", err)
	}

	return p, nil
}

func (p *Producer) ping(ctx context.Context) error {
	conn, err := kafka.DialContext(ctx, "tcp", p.brokers[0])
	if err != nil {
		return fmt.Errorf("Kafka Producer - kafka.DialContext: %w", err)
	}
	defer conn.Close()

	_, err = conn.Brokers()
	if err != nil {
		return fmt.Errorf("Kafka Producer - conn.Brokers: %w", err)
	}

	return nil
}

func (p *Producer) Close() error {
	if p.Writer != nil {
		return p.Writer.Close()
	}

	return nil
}
