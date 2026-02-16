package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type (
	Config struct {
		HTTP            HTTP
		Log             Log
		PG              PG
		S3              S3
		OutboxRelay     OutboxRelay
		Kafka           Kafka
		KafkaController KafkaController
		Swagger         Swagger
	}

	HTTP struct {
		Port           string `env:"HTTP_PORT,required"`
		UsePreforkMode bool   `env:"HTTP_USE_PREFORK_MODE" envDefault:"false"`
	}

	Log struct {
		Level string `env:"LOG_LEVEL,required"`
	}

	PG struct {
		PoolMax int    `env:"PG_POOL_MAX,required"`
		URL     string `env:"PG_URL,required"`
	}

	S3 struct {
		Endpoint       string        `env:"S3_ENDPOINT,required"`
		AccessKey      string        `env:"S3_ACCESS_KEY,required"`
		SecretKey      string        `env:"S3_SECRET_KEY,required"`
		Bucket         string        `env:"S3_BUCKET,required"`
		CfgLoadTimeout time.Duration `env:"S3_LOAD_CFG_TIMEOUT" envDefault:"10s"`
	}

	Kafka struct {
		Brokers []string `env:"KAFKA_BROKERS,required"`
		GroupID string   `env:"KAFKA_GROUP_ID,required"`
		Topic   string   `env:"KAFKA_TOPIC,required"`
	}

	OutboxRelay struct {
		PollInterval        time.Duration `env:"OUTBOX_RELAY_POLL_INTERVAL" envDefault:"2s"`
		MarkFailedInterval  time.Duration `env:"OUTBOX_RELAY_MARK_FAILED_INTERVAL" envDefault:"2m"`
		CleanupInterval     time.Duration `env:"OUTBOX_RELAY_CLEANUP_INTERVAL" envDefault:"24h"`
		ProcessBatchTimeout time.Duration `env:"OUTBOX_RELAY_PROCESS_BATCH_TIMEOUT" envDefault:"15s"`
		ShutdownTimeout     time.Duration `env:"OUTBOX_RELAY_SHUTDOWN_TIMEOUT" envDefault:"5s"`
		BatchSize           int           `env:"OUTBOX_RELAY_BATCH_SIZE" envDefault:"100"`
		MaxRetries          int           `env:"OUTBOX_RELAY_MAX_RETRIES" envDefault:"3"`
	}

	KafkaController struct {
		CommitTimeout   time.Duration `env:"KAFKA_CONTROLLER_COMMIT_TIMEOUT" envDefault:"2s"`
		ProcessTimeout  time.Duration `env:"KAFKA_CONTROLLER_PROCESS_TIMEOUT" envDefault:"15s"` // вся операция - чтение/запись в хранилище и БД, обработка изображения
		CPUTimeout      time.Duration `env:"KAFKA_CONTROLLER_CPU_TIMEOUT" envDefault:"8s"`      // обработка изображения
		ShutdownTimeout time.Duration `env:"KAFKA_CONTROLLER_SHUTDOWN_TIMEOUT" envDefault:"5s"`
	}

	Swagger struct {
		Enabled bool `env:"SWAGGER_ENABLED" envDefault:"false"`
	}
)

func New() (*Config, error) {
	cfg := &Config{}

	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	return cfg, nil
}
