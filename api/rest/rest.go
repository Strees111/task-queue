package rest

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"projectgo/api/core"
)

func NewHealthcheckHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"OK"}`))
	})
}

func NewEnqueueHandler(log *slog.Logger, wp core.WorkerPool, runner core.ExponentialBackoffWithJitter) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var task core.Task
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			writeJSONError(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if task.Id == "" || task.Max_retries <= 0 {
			writeJSONError(w, "Invalid task parameters", http.StatusBadRequest)
			return
		}

		err := wp.Submit(r.Context(), func() {
			task.Status = "queued"
			log.Info("Task queued", "task_id", task.Id, "status", task.Status)

			_ = runner.ProcessTask(context.Background(), &task)
		})

		if err != nil {
			log.Error("Failed to submit task", "error", err)

			if errors.Is(err, core.ErrPoolFull) {
				writeJSONError(w, "Queue is full, try again later", http.StatusServiceUnavailable)
				return
			}
			if errors.Is(err, core.ErrPoolClosed) {
				writeJSONError(w, "Server is shutting down", http.StatusServiceUnavailable)
				return
			}
			if errors.Is(err, context.Canceled) {
				return
			}

			writeJSONError(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		if err := json.NewEncoder(w).Encode(task); err != nil {
			log.ErrorContext(r.Context(), "failed to encode response", "error", err)
		}
	})
}

func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := struct {
		Error string `json:"error"`
	}{
		Error: message,
	}

	_ = json.NewEncoder(w).Encode(response)
}
