package outbox

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/andreyxaxa/Image-Processor/internal/infrastructure"
	"github.com/andreyxaxa/Image-Processor/internal/usecase"
	"github.com/andreyxaxa/Image-Processor/pkg/logger"
)

type OutboxRelay struct {
	img    usecase.ImageUseCase
	es     infrastructure.EventsSender
	logger logger.Interface

	pollInterval        time.Duration
	cleanupInterval     time.Duration
	markFailedInterval  time.Duration
	processBatchTimeout time.Duration
	batchSize           int
	maxRetries          int

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	started atomic.Bool
}

func New(
	img usecase.ImageUseCase,
	es infrastructure.EventsSender,
	l logger.Interface,
	pollInterval time.Duration,
	cleanupInterval time.Duration,
	markFailedInterval time.Duration,
	processBatchTimeout time.Duration,
	batchSize int,
	maxRetries int,
) *OutboxRelay {
	return &OutboxRelay{
		img:                 img,
		es:                  es,
		logger:              l,
		pollInterval:        pollInterval,
		cleanupInterval:     cleanupInterval,
		markFailedInterval:  markFailedInterval,
		processBatchTimeout: processBatchTimeout,
		batchSize:           batchSize,
		maxRetries:          maxRetries,
	}
}

func (r *OutboxRelay) Start(ctx context.Context) error {
	if !r.started.CompareAndSwap(false, true) {
		return fmt.Errorf("OutboxRelay - Run - worker already started")
	}

	r.ctx, r.cancel = context.WithCancel(ctx)

	// 1. воркер для отправки задач в очередь
	r.worker(r.pollInterval, func() {
		batchCtx, batchCancel := context.WithTimeout(r.ctx, r.processBatchTimeout)
		r.processEventsBatch(batchCtx)
		batchCancel()
	})

	// 2. воркер для пометки failed
	r.worker(r.markFailedInterval, func() {
		err := r.img.MarkMaxRetriesAsFailed(r.ctx, r.maxRetries)
		if err != nil {
			r.logger.Error(err, "OutboxRelay - Start - worker - r.img.MarkMaxRetriesAsFailed")
		}
	})

	// 3. воркер очистки failed/processed из outbox
	r.worker(r.cleanupInterval, func() {
		err := r.img.CleanupOutbox(r.ctx)
		if err != nil {
			r.logger.Error(err, "OutboxRelay - Start - worker - r.img.CleanupOutbox")
		}
	})

	return nil
}

func (r *OutboxRelay) processEventsBatch(ctx context.Context) {
	// 1. получаем events со статусом pending, у которых retry count < max retries
	events, err := r.img.GetPendingEvents(ctx, r.maxRetries, r.batchSize)
	if err != nil {
		r.logger.Error(err, "OutboxRelay - processEventsBatch - r.img.GetEventsToSend")

		return
	}
	if len(events) == 0 {
		return
	}

	// 2. помечаем как processing
	err = r.img.MarkAsProcessingBatch(ctx, events)
	if err != nil {
		r.logger.Error(err, "OutboxRelay - processEventsBatch - r.img.MarkAsProcessingBatch")

		return
	}

	// 3. пробуем их отправить
	err = r.es.SendEvents(ctx, events)
	if err != nil {
		r.logger.Error(err, "OutboxRelay - processEventsBatch - r.es.SendEvents")
		// 3.1 если не получилось - увеличиваем счетчик ретраев + возвращаем статус в pending
		incErr := r.img.IncrementRetryCountBatch(ctx, events)
		if incErr != nil {
			r.logger.Error(incErr, "OutboxRelay - processEventsBatch - r.img.IncrementRetryCountBatch")
		}
		return
	}

	// 4. если удалось отправить - помечаем как processed
	err = r.img.MarkAsProcessedBatch(ctx, events)
	if err != nil {
		r.logger.Error(err, "OutboxRelay - processEventsBatch - r.img.MarkAsProcessedBatch")

		return
	}
}

func (r *OutboxRelay) worker(interval time.Duration, task func()) {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-r.ctx.Done():
				return
			case <-ticker.C:
				task()
			}
		}
	}()
}

func (r *OutboxRelay) Shutdown(ctx context.Context) error {
	if !r.started.Load() {
		return nil
	}

	if r.cancel != nil {
		r.cancel()
	}

	done := make(chan struct{})

	go func() {
		r.wg.Wait()
		r.es.Close()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return nil
	}
}
