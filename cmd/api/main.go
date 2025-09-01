package main

import (
	"distributed-job-queue/internal/api"
	"distributed-job-queue/internal/db"
	"distributed-job-queue/internal/queue"
	"distributed-job-queue/pkg/config"
	"log"
	"net/http"
)

func main() {
	cfg := config.LoadConfig()

	q := queue.NewRedisQueue(cfg.RedisAddr, cfg.QueueName)
	database, err := db.NewDB(cfg.PostgresURL)
	if err != nil {
		log.Fatal("DB connection failed:", err)
	}

	apiServer := api.NewAPI(database, q)

	log.Println("API server running on :8080")
	http.ListenAndServe(":8080", apiServer.Router())
}
