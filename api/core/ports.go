package core

import (
	"context"
)

type WorkerPool interface {
	Submit(ctx context.Context, task func()) error
	Stop(ctx context.Context) error
}

type ExponentialBackoffWithJitter interface {
	ProcessTask(ctx context.Context, task *Task) error
}
