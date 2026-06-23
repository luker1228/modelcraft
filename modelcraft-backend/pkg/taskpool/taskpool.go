// Package taskpool provides a bounded worker pool for background tasks.
//
// It wraps github.com/alitto/pond with project-standard panic recovery and
// structured logging via logfacade. Tasks are named so that operators can
// attribute duration, panic, and failure metrics to a specific business flow.
package taskpool

import (
	"context"
	"errors"
	"modelcraft/pkg/logfacade"
	"time"

	"github.com/alitto/pond"
)

// TaskPool is a bounded worker pool with a fixed number of workers and a
// bounded queue. Submit is non-blocking: when the queue is full it returns
// ErrQueueFull immediately so callers can degrade gracefully (e.g. mark the
// job as failed) instead of blocking the request goroutine.
type TaskPool struct {
	pool *pond.WorkerPool
}

// NewTaskPool constructs a TaskPool with the given workerNum (concurrent
// workers) and queueSize (pending task buffer). Panics if workerNum <= 0.
func NewTaskPool(workerNum, queueSize int) *TaskPool {
	if workerNum <= 0 {
		panic("taskpool: workerNum must be > 0")
	}
	return &TaskPool{
		pool: pond.New(workerNum, queueSize),
	}
}

// ErrQueueFull is returned by Submit when the pending task queue is full.
var ErrQueueFull = errors.New("task pool queue is full")

// Submit enqueues a named task. It returns ErrQueueFull immediately if the
// queue is full; otherwise it returns nil and the task runs asynchronously.
//
// Inside the worker, panics are recovered and logged via logfacade, task
// errors are logged, and task duration is measured. Metric emission points
// are marked with comments so prometheus collectors can be wired in later
// without touching call sites.
func (p *TaskPool) Submit(name string, fn func() error) error {
	ok := p.pool.TrySubmit(func() {
		start := time.Now()

		defer func() {
			cost := time.Since(start)

			// metric: taskDuration.WithLabelValues(name).Observe(cost.Seconds())
			_ = cost

			if r := recover(); r != nil {
				// metric: taskPanicCounter.WithLabelValues(name).Inc()
				logfacade.GetLogger(context.Background()).With(
					logfacade.String("task", name),
					logfacade.Any("panic_value", r),
				).Errorf(context.Background(), nil, "task panic: %s", name)
			}
		}()

		if err := fn(); err != nil {
			// metric: taskFailedCounter.WithLabelValues(name).Inc()
			logfacade.GetLogger(context.Background()).With(
				logfacade.String("task", name),
				logfacade.Err(err),
			).Errorf(context.Background(), nil, "task failed: %s", name)
			return
		}

		// metric: taskSuccessCounter.WithLabelValues(name).Inc()
	})

	if !ok {
		// metric: taskRejectedCounter.WithLabelValues(name).Inc()
		return ErrQueueFull
	}
	return nil
}

// Close stops the pool and waits for all in-flight tasks to finish.
func (p *TaskPool) Close() {
	p.pool.StopAndWait()
}
