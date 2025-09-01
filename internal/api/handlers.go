package api

import (
	"distributed-job-queue/internal/db"
	"distributed-job-queue/internal/models"
	"distributed-job-queue/internal/queue"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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
	mux.HandleFunc("/status/", a.StatusHandler)
	return mux
}

func (a *API) EnqueueHandler(w http.ResponseWriter, r *http.Request) {
	var payload map[string]string
	json.NewDecoder(r.Body).Decode(&payload)

	job := models.Job{
		ID:      uuid.New().String(),
		Type:    "mock",
		Payload: fmt.Sprintf(`"%s"`, payload["data"]),
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

func (a *API) StatusHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	prefix := "/status/"
	if !strings.HasPrefix(path, prefix) {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	id := strings.TrimPrefix(path, prefix)
	if id == "" {
		http.Error(w, "missing job id", http.StatusBadRequest)
		return
	}

	job, err := a.db.GetJob(r.Context(), id)
	if err != nil {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}
