package producer

import "time"

type Option func(*Producer)

func ConnAttempts(attempts int) Option {
	return func(p *Producer) {
		p.connAttempts = attempts
	}
}

func ConnTimeout(timeout time.Duration) Option {
	return func(p *Producer) {
		p.connTimeout = timeout
	}
}
