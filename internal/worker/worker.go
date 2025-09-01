package worker

import (
	"context"
	"distributed-job-queue/internal/db"
	"distributed-job-queue/internal/models"
	"distributed-job-queue/internal/queue"
	"fmt"
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

		w.db.UpdateJobStatus(ctx, job.ID, models.StatusSuccess)
	}
}
