package worker

import (
	"context"
	"distributed-job-queue/internal/db"
	"distributed-job-queue/internal/models"
	"distributed-job-queue/internal/queue"
	"encoding/json"
	"fmt"
	"time"
)

type Worker struct {
	db    *db.DB
	queue *queue.RedisQueue
}

func NewWorker(db *db.DB, q *queue.RedisQueue) *Worker {
	return &Worker{db: db, queue: q}
}

func (w *Worker) Start(ctx context.Context) {
	for {
		jobID, err := w.queue.Dequeue(ctx)
		if err != nil {
			fmt.Println("Queue error:", err)
			continue
		}

		job, err := w.db.GetJob(ctx, jobID)
		if err != nil {
			fmt.Println("DB error:", err)
			continue
		}

		fmt.Printf("Dequeued job %s: retries=%d, max_retries=%d\n", job.ID, job.Retries, job.MaxRetries)
		w.db.UpdateJobStatus(ctx, job.ID, models.StatusRunning)

		// Give each job a max execution window
		execCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		err = w.execute(execCtx, job)
		cancel()

		if err != nil {
			job.Retries++
			job.ErrorMessage = err.Error()

			if job.Retries >= job.MaxRetries {
				fmt.Println("Job failed permanently:", job.ID, "error:", job.ErrorMessage)
				w.db.UpdateJobFailure(ctx, job.ID, job.Retries, job.MaxRetries, job.ErrorMessage)
			} else {
				fmt.Println("Retrying job:", job.ID, "attempt", job.Retries)
				w.db.UpdateJobRetry(ctx, job.ID, job.Retries, job.ErrorMessage) // ✅ persist retries
				time.Sleep(time.Duration(job.Retries) * time.Second)            // backoff
				w.db.UpdateJobStatus(ctx, job.ID, models.StatusPending)
				w.queue.Enqueue(ctx, job.ID)
			}
			continue
		}

		// Success
		fmt.Println("Job succeeded:", job.ID)
		w.db.UpdateJobStatus(ctx, job.ID, models.StatusSuccess)
	}
}

// Execute job based on payload content
func (w *Worker) execute(ctx context.Context, job models.Job) error {
	var payload string
	if err := json.Unmarshal([]byte(job.Payload), &payload); err != nil {
		return fmt.Errorf("invalid payload: %v", err)
	}

	switch payload {
	case "fail":
		return fmt.Errorf("simulated failure for payload: %s", payload)

	case "slow":
		// Simulate long-running job that exceeds timeout
		select {
		case <-time.After(5 * time.Second): // job takes 5s
			return nil
		case <-ctx.Done():
			return fmt.Errorf("job timed out: %v", ctx.Err())
		}

	default:
		// Normal quick job
		select {
		case <-time.After(1 * time.Second):
			return nil
		case <-ctx.Done():
			return fmt.Errorf("job timed out: %v", ctx.Err())
		}
	}
}
