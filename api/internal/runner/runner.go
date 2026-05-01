package runner

import (
	"context"
	"log/slog"
	"math"
	"math/rand/v2"
	"time"

	"projectgo/api/core"
)

type Runner struct {
	log *slog.Logger
}

func NewRunner(log *slog.Logger) *Runner {
	return &Runner{
		log: log,
	}
}

func (r *Runner) ProcessTask(ctx context.Context, task *core.Task) error {
	baseDelay := 100 * time.Millisecond
	maxDelay := 5 * time.Second

	var err error

	for attempt := 1; attempt <= task.Max_retries; attempt++ {
		err = r.doWork(ctx, task)
		if err == nil {
			return nil
		}

		r.log.Warn("Task failed, retrying",
			"task_id", task.Id,
			"attempt", attempt,
			"max_retries", task.Max_retries,
			"error", err,
		)

		if attempt == task.Max_retries {
			break
		}

		backoffTime := baseDelay * time.Duration(math.Pow(2, float64(attempt-1)))
		if backoffTime > maxDelay {
			backoffTime = maxDelay
		}
		jitter := time.Duration(rand.Int64N(int64(backoffTime)))

		timer := time.NewTimer(jitter)
		select {
		case <-ctx.Done():
			timer.Stop()
			r.log.Info("Task processing cancelled by context", "task_id", task.Id)
			return ctx.Err()
		case <-timer.C:
		}
		
	}

	r.log.Error("Task completely failed after retries", "task_id", task.Id, "error", err)
	return err
}

func (r *Runner) doWork(ctx context.Context, task *core.Task) error {
	task.Status = "running"
	r.log.Debug("Task running", "task_id", task.Id, "status", task.Status)

	workTime := 100 + rand.IntN(400)

	timer := time.NewTimer(time.Duration(workTime) * time.Millisecond)
	select {
	case <-ctx.Done():
		timer.Stop()
		task.Status = "cancelled"
		return ctx.Err()
	case <-timer.C:
	}

	if rand.IntN(10) < 2 {
		task.Status = "failed"
		r.log.Error("Task encountered temporary failure", "task_id", task.Id)
		return core.ErrTaskFailed
	}

	task.Status = "done"
	r.log.Info("Task completed successfully", "task_id", task.Id, "status", task.Status)
	return nil
}
