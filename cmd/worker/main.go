package main

import (
	"context"
	"distributed-job-queue/internal/db"
	"distributed-job-queue/internal/queue"
	"distributed-job-queue/internal/worker"
	"distributed-job-queue/pkg/config"
	"log"
)

func main() {
	cfg := config.LoadConfig()

	q := queue.NewRedisQueue(cfg.RedisAddr, cfg.QueueName)
	database, err := db.NewDB(cfg.PostgresURL)
	if err != nil {
		log.Fatal("DB connection failed:", err)
	}

	w := worker.NewWorker(database, q)
	w.Start(context.Background())
}
