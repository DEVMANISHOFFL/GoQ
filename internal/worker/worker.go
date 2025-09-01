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

		w.db.UpdateJobStatus(ctx, job.ID, models.StatusRunning)

		fmt.Println("Processing job:", job.ID, "Payload:", job.Payload)

		err = w.execute(job)
		if err != nil {
			job.Retries++
			job.ErrorMessage = err.Error()

			if job.Retries >= job.MaxRetries {
				fmt.Println("Job failed permanently:", job.ID)
				w.db.UpdateJobFailure(ctx, job.ID, job.Retries, job.MaxRetries, job.ErrorMessage)
			} else {
				fmt.Println("Retrying job:", job.ID, "attempt", job.Retries)
				time.Sleep(time.Duration(job.Retries) * time.Second)
				w.db.UpdateJobStatus(ctx, job.ID, models.StatusPending)
				w.db.SaveJob(ctx, job)
				w.queue.Enqueue(ctx, job.ID)
			}
			continue
		}

		w.db.UpdateJobStatus(ctx, job.ID, models.StatusSuccess)
	}
}

func (w *Worker) execute(job models.Job) error {
	var data string
	if err := json.Unmarshal([]byte(job.Payload), &data); err != nil {
		return fmt.Errorf("invalid payload: %v", err)
	}

	if data == "fail" {
		return fmt.Errorf("simulated failure for payload: %s", data)
	}
	return nil
}
