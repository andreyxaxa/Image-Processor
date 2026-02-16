package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/andreyxaxa/Image-Processor/config"
	kafkactrl "github.com/andreyxaxa/Image-Processor/internal/controller/kafka"
	"github.com/andreyxaxa/Image-Processor/internal/controller/restapi"
	"github.com/andreyxaxa/Image-Processor/internal/controller/worker/outbox"
	infrakafka "github.com/andreyxaxa/Image-Processor/internal/infrastructure/kafka"
	"github.com/andreyxaxa/Image-Processor/internal/infrastructure/processor"
	"github.com/andreyxaxa/Image-Processor/internal/repo/persistent"
	"github.com/andreyxaxa/Image-Processor/internal/usecase/image"
	"github.com/andreyxaxa/Image-Processor/internal/usecase/imageprocessor"
	"github.com/andreyxaxa/Image-Processor/pkg/httpserver"
	"github.com/andreyxaxa/Image-Processor/pkg/kafka/consumer"
	"github.com/andreyxaxa/Image-Processor/pkg/kafka/producer"
	"github.com/andreyxaxa/Image-Processor/pkg/logger"
	"github.com/andreyxaxa/Image-Processor/pkg/postgres"
	"github.com/andreyxaxa/Image-Processor/pkg/s3client"
)

func Run(cfg *config.Config) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Logger
	l := logger.New(cfg.Log.Level)

	// Repository

	// s3
	s3Ctx, s3Cancel := context.WithTimeout(ctx, cfg.S3.CfgLoadTimeout)
	defer s3Cancel()
	s3c, err := s3client.New(s3Ctx, cfg.S3.Endpoint, cfg.S3.AccessKey, cfg.S3.SecretKey)
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - s3client.New: %w", err))
	}

	// postgres
	pg, err := postgres.New(cfg.PG.URL, postgres.MaxPoolSize(cfg.PG.PoolMax))
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - postgres.New: %w", err))
	}
	defer pg.Close()

	// Use-Case

	// image use-case
	imageUseCase := image.New(
		persistent.NewImageRepo(s3c, cfg.S3.Bucket),
		persistent.NewImageMetadataRepo(pg),
		persistent.NewOutboxImageMetadataRepo(pg),
		pg,
		l,
	)

	// image processor use-case
	imageProcessorUseCase := imageprocessor.New(processor.New())

	// Kafka Producer
	kafkaProducer, err := producer.New(ctx, cfg.Kafka.Brokers)
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - producer.New: %w", err))
	}

	// Outbox Relay Worker
	outboxRelayWorker := outbox.New(
		imageUseCase,
		infrakafka.NewEventProducer(kafkaProducer, cfg.OutboxRelay.MaxRetries, cfg.Kafka.Topic),
		l,
		cfg.OutboxRelay.PollInterval,
		cfg.OutboxRelay.CleanupInterval,
		cfg.OutboxRelay.MarkFailedInterval,
		cfg.OutboxRelay.ProcessBatchTimeout,
		cfg.OutboxRelay.BatchSize,
		cfg.OutboxRelay.MaxRetries,
	)

	// Kafka Consumer
	kafkaConsumer, err := consumer.New(ctx, cfg.Kafka.Brokers, cfg.Kafka.GroupID, cfg.Kafka.Topic)
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - consumer.New: %w", err))
	}

	// Kafka as Controller
	kafkaController := kafkactrl.New(
		imageProcessorUseCase,
		imageUseCase,
		infrakafka.NewEventConsumer(kafkaConsumer),
		l,
		cfg.KafkaController.CommitTimeout,
		cfg.KafkaController.ProcessTimeout,
		cfg.KafkaController.CPUTimeout,
		runtime.NumCPU(),
	)

	// HTTP Server
	httpServer := httpserver.New(l, httpserver.Port(cfg.HTTP.Port), httpserver.Prefork(cfg.HTTP.UsePreforkMode))
	restapi.NewRouter(httpServer.App, cfg, imageUseCase, l)

	// Start Components
	err = outboxRelayWorker.Start(ctx)
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - outboxRelayWorker.Start: %w", err))
	}
	err = kafkaController.Start(ctx)
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - kafkaController.Start: %w", err))
	}
	httpServer.Start()

	// Waiting Signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		l.Info("app - Run - signal: %s", s.String())
	case err = <-httpServer.Notify():
		l.Error(fmt.Errorf("app - Run - httpServer.Notify: %w", err))
	}

	// Shutdown
	err = httpServer.Shutdown()
	if err != nil {
		l.Error(fmt.Errorf("app - Run - httpServer.Shutdown: %w", err))
	}

	orlShutdownCtx, orlShutdownCancel := context.WithTimeout(ctx, cfg.OutboxRelay.ShutdownTimeout)
	defer orlShutdownCancel()
	err = outboxRelayWorker.Shutdown(orlShutdownCtx)
	if err != nil {
		l.Error(fmt.Errorf("app - Run - outboxRelayWorker.Shutdown: %w", err))
	}

	kcShutdownCtx, kcShutdownCancel := context.WithTimeout(ctx, cfg.KafkaController.ShutdownTimeout)
	defer kcShutdownCancel()
	err = kafkaController.Shutdown(kcShutdownCtx)
	if err != nil {
		l.Error(fmt.Errorf("app - Run - kafkaController.Shutdown: %w", err))
	}
}
