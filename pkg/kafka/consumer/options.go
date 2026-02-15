package consumer

import "time"

type Option func(*Consumer)

func ConnAttempts(attempts int) Option {
	return func(c *Consumer) {
		c.connAttempts = attempts
	}
}

func ConnTimeout(timeout time.Duration) Option {
	return func(c *Consumer) {
		c.connTimeout = timeout
	}
}
