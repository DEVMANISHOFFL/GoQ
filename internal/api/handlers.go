package api

import (
	"distributed-job-queue/internal/db"
	"distributed-job-queue/internal/models"
	"distributed-job-queue/internal/queue"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type API struct {
	db    *db.DB
	queue *queue.RedisQueue
}

func NewAPI(db *db.DB, q *queue.RedisQueue) *API {
	return &API{db: db, queue: q}
}

func (a *API) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/enqueue", a.EnqueueHandler)
	return mux
}

func (a *API) EnqueueHandler(w http.ResponseWriter, r *http.Request) {
	var payload map[string]string
	json.NewDecoder(r.Body).Decode(&payload)

	job := models.Job{
		ID:      uuid.New().String(),
		Type:    "mock",
		Payload: payload["data"],
		Status:  models.StatusPending,
	}

	if err := a.db.SaveJob(r.Context(), job); err != nil {
		http.Error(w, "failed to save job", http.StatusInternalServerError)
		return
	}

	if err := a.queue.Enqueue(r.Context(), job.ID); err != nil {
		http.Error(w, "failed to enqueue job", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(job)
}
