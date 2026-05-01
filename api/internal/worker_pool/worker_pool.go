package worker_pool

import (
	"context"
	"projectgo/api/core"
	"sync"
)

type WorkerPool struct {
	mu       sync.RWMutex
	isClosed bool
	tasks    chan func()
	wg       sync.WaitGroup
}

func NewWorkerPool(workersCount, queueSize int) *WorkerPool {
	if workersCount <= 0 {
		workersCount = 1
	}
	if queueSize <= 0 {
		queueSize = 64
	}

	wp := &WorkerPool{
		tasks: make(chan func(), queueSize),
	}

	wp.wg.Add(workersCount)
	for i := 0; i < workersCount; i++ {
		go wp.worker()
	}

	return wp
}

func (wp *WorkerPool) worker() {
	defer wp.wg.Done()

	for task := range wp.tasks {
		if task != nil {
			task()
		}
	}
}

func (wp *WorkerPool) Submit(ctx context.Context, task func()) error {
	if task == nil {
		return core.ErrInvalidTask
	}

	wp.mu.RLock()
	isClosed := wp.isClosed
	wp.mu.RUnlock()

	if isClosed {
		return core.ErrPoolClosed
	}

	select {
	case wp.tasks <- task:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return core.ErrPoolFull
	}
}

func (wp *WorkerPool) Stop(ctx context.Context) error {
	wp.mu.Lock()
	if wp.isClosed {
		wp.mu.Unlock()
		return core.ErrPoolClosed
	}
	wp.isClosed = true
	close(wp.tasks)
	wp.mu.Unlock()

	done := make(chan struct{})
	go func() {
		wp.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
