package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/andreyxaxa/Image-Processor/internal/dto"
	kafkapc "github.com/andreyxaxa/Image-Processor/internal/infrastructure/kafka"
	"github.com/andreyxaxa/Image-Processor/internal/usecase"
	"github.com/andreyxaxa/Image-Processor/pkg/logger"
	"github.com/segmentio/kafka-go"
)

type KafkaController struct {
	prc    usecase.ImageProcessorUseCase
	img    usecase.ImageUseCase
	ec     *kafkapc.EventConsumer
	logger logger.Interface

	commitTimeout  time.Duration
	processTimeout time.Duration
	cpuTimeout     time.Duration

	workers int
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup

	started atomic.Bool
}

func New(
	p usecase.ImageProcessorUseCase,
	img usecase.ImageUseCase,
	ec *kafkapc.EventConsumer,
	l logger.Interface,
	commitTimeout time.Duration,
	processTimeout time.Duration,
	cpuTimeout time.Duration,
	workers int,
) *KafkaController {
	return &KafkaController{
		prc:            p,
		img:            img,
		ec:             ec,
		logger:         l,
		commitTimeout:  commitTimeout,
		processTimeout: processTimeout,
		cpuTimeout:     cpuTimeout,
		workers:        workers,
	}
}

func (c *KafkaController) Start(ctx context.Context) error {
	if !c.started.CompareAndSwap(false, true) {
		return fmt.Errorf("KafkaController - Start - controller already started")
	}

	c.ctx, c.cancel = context.WithCancel(ctx)

	// канал для задач
	tasks := make(chan kafka.Message, c.workers*2)

	// запускаем воркеры
	for i := 0; i < c.workers; i++ {
		c.wg.Add(1)
		go c.worker(tasks)
	}

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer close(tasks)

		for {
			select {
			case <-c.ctx.Done():
				return
			default:
				// 1. читаем из кафки
				event, err := c.ec.ReadEvent(c.ctx)
				if err != nil {
					if !errors.Is(err, context.Canceled) {
						c.logger.Error(err, "KafkaController - Start - c.ec.ReadEvent")
					}
					continue
				}

				// 2. отправляем в канал для воркеров
				select {
				case tasks <- event:
				case <-c.ctx.Done():
					return
				}
			}
		}
	}()

	return nil
}

func (c *KafkaController) processImage(ctx context.Context, event kafka.Message) error {
	var payload ImageEventPayload
	err := json.Unmarshal(event.Value, &payload)
	if err != nil {
		return fmt.Errorf("KafkaController - processImage - json.Unmarshal: %w", err)
	}

	// 1. скачиваем из S3
	data, err := c.img.DownloadImageBytes(ctx, payload.OriginalKey)
	if err != nil {
		return fmt.Errorf("KafkaController - processImage - c.img.DownloadImageBytes: %w", err)
	}

	// 2. формируем dto, обрабатываем
	cpuCtx, cpuCancel := context.WithTimeout(ctx, c.cpuTimeout)
	defer cpuCancel()
	processed, err := c.prc.Process(cpuCtx, payload.ContentType, dto.Task{
		Data:      data,
		Operation: payload.Operation,
		Width:     payload.Width,
		Height:    payload.Height,
	})
	if err != nil {
		return fmt.Errorf("KafkaController - processImage - c.prc.Process: %w", err)
	}

	// 3. загружаем в S3 обработанное изображение, обновляем метаданные в бд
	err = c.img.UploadProcessedImage(ctx, processed, payload.ID)
	if err != nil {
		return fmt.Errorf("KafkaController - processImage - c.img.UploadProcessedImage: %w", err)
	}

	return nil
}

func (c *KafkaController) worker(tasks <-chan kafka.Message) {
	defer c.wg.Done()

	// читаем канал, пока не закроется
	for event := range tasks {
		func() {
			defer func() {
				if r := recover(); r != nil {
					c.logger.Error(fmt.Errorf("panic %v", r), "KafkaController - worker - panic")
				}
			}()

			// выполняем обработку
			processCtx, processCancel := context.WithTimeout(c.ctx, c.processTimeout)
			err := c.processImage(processCtx, event)
			processCancel()
			if err != nil {
				c.logger.Error(err, "KafkaController - worker - c.processImage: %w", err)

				return
			}

			// коммитим после успешной обработки
			commitCtx, commitCancel := context.WithTimeout(c.ctx, c.commitTimeout)
			err = c.ec.CommitEvent(commitCtx, event)
			commitCancel()
			if err != nil {
				c.logger.Error(err, "KafkaController - worker - c.ec.CommitEvent: %w", err)
			}
		}()
	}
}

func (c *KafkaController) Shutdown(ctx context.Context) error {
	if !c.started.Load() {
		return nil
	}

	if c.cancel != nil {
		c.cancel()
	}

	done := make(chan struct{})

	go func() {
		c.wg.Wait()
		c.ec.Close()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return nil
	}
}
