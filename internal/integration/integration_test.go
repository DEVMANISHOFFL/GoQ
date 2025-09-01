package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"distributed-job-queue/internal/api"
	"distributed-job-queue/internal/db"
	"distributed-job-queue/internal/models"
	"distributed-job-queue/internal/queue"
	"distributed-job-queue/internal/worker"
)

func pollJobStatus(t *testing.T, serverURL, jobID string, deadline time.Duration) models.Job {
	t.Helper()

	var job models.Job
	timeout := time.Now().Add(deadline)

	for {
		if time.Now().After(timeout) {
			t.Fatalf("job %s did not complete in time", jobID)
		}

		res, err := http.Get(serverURL + "/status/" + jobID)
		if err != nil {
			t.Fatalf("status request failed: %v", err)
		}
		defer res.Body.Close()

		if err := json.NewDecoder(res.Body).Decode(&job); err != nil {
			t.Fatalf("decode status resp: %v", err)
		}

		if job.Status == models.StatusSuccess || job.Status == models.StatusFailed {
			return job
		}

		time.Sleep(500 * time.Millisecond)
	}
}

func TestJobLifecycle(t *testing.T) {
	ctx := context.Background()

	d, err := db.NewDB("postgres://user:password@localhost:5432/jobs?sslmode=disable")
	if err != nil {
		t.Fatalf("failed to connect db: %v", err)
	}
	q := queue.NewRedisQueue("localhost:6379", "jobs")

	apiSvc := api.NewAPI(d, q)
	server := httptest.NewServer(apiSvc.Router())
	defer server.Close()

	w := worker.NewWorker(d, q)
	go func() {
		w.Start(ctx)
	}()

	t.Run("successful job", func(t *testing.T) {
		payload := []byte(`{"data":"send_email"}`)
		resp, err := http.Post(server.URL+"/enqueue", "application/json", bytes.NewBuffer(payload))
		if err != nil {
			t.Fatalf("enqueue request failed: %v", err)
		}
		defer resp.Body.Close()

		var enqResp models.Job
		if err := json.NewDecoder(resp.Body).Decode(&enqResp); err != nil {
			t.Fatalf("decode enqueue resp: %v", err)
		}

		job := pollJobStatus(t, server.URL, enqResp.ID, 5*time.Second)
		if job.Status != models.StatusSuccess {
			t.Fatalf("expected success, got %s", job.Status)
		}
	})

	t.Run("failing job with retries", func(t *testing.T) {
		payload := []byte(`{"data":"fail"}`)
		resp, err := http.Post(server.URL+"/enqueue", "application/json", bytes.NewBuffer(payload))
		if err != nil {
			t.Fatalf("enqueue request failed: %v", err)
		}
		defer resp.Body.Close()

		var enqResp models.Job
		if err := json.NewDecoder(resp.Body).Decode(&enqResp); err != nil {
			t.Fatalf("decode enqueue resp: %v", err)
		}

		job := pollJobStatus(t, server.URL, enqResp.ID, 15*time.Second)
		if job.Status != models.StatusFailed {
			t.Fatalf("expected failed, got %s", job.Status)
		}
		if job.ErrorMessage == "" {
			t.Fatalf("expected error message, got empty string")
		}
	})

	t.Run("timeout job", func(t *testing.T) {
		payload := []byte(`{"data":"slow"}`)
		resp, err := http.Post(server.URL+"/enqueue", "application/json", bytes.NewBuffer(payload))
		if err != nil {
			t.Fatalf("enqueue request failed: %v", err)
		}
		defer resp.Body.Close()

		var enqResp models.Job
		if err := json.NewDecoder(resp.Body).Decode(&enqResp); err != nil {
			t.Fatalf("decode enqueue resp: %v", err)
		}

		job := pollJobStatus(t, server.URL, enqResp.ID, 20*time.Second)
		if job.Status != models.StatusFailed {
			t.Fatalf("expected failed (timeout), got %s", job.Status)
		}
		if job.ErrorMessage == "" || !strings.HasPrefix(job.ErrorMessage, "job timed out") {
			t.Fatalf("expected timeout error, got %s", job.ErrorMessage)
		}

	})

	t.Run("mixed batch of jobs", func(t *testing.T) {
		jobs := []string{"send_email", "fail", "slow"}
		var jobIDs []string

		for _, payloadVal := range jobs {
			payload := []byte(`{"data":"` + payloadVal + `"}`)
			resp, err := http.Post(server.URL+"/enqueue", "application/json", bytes.NewBuffer(payload))
			if err != nil {
				t.Fatalf("enqueue request failed: %v", err)
			}
			defer resp.Body.Close()

			var enqResp models.Job
			if err := json.NewDecoder(resp.Body).Decode(&enqResp); err != nil {
				t.Fatalf("decode enqueue resp: %v", err)
			}
			jobIDs = append(jobIDs, enqResp.ID)
		}

		for i, id := range jobIDs {
			job := pollJobStatus(t, server.URL, id, 20*time.Second)
			switch jobs[i] {
			case "send_email":
				if job.Status != models.StatusSuccess {
					t.Fatalf("expected success, got %s for job %s", job.Status, id)
				}
			case "fail":
				if job.Status != models.StatusFailed {
					t.Fatalf("expected failed, got %s for job %s", job.Status, id)
				}
			case "slow":
				if job.Status != models.StatusFailed {
					t.Fatalf("expected timeout fail, got %s for job %s", job.Status, id)
				}
				if !strings.HasPrefix(job.ErrorMessage, "job timed out") {
					t.Fatalf("expected timeout error, got %s", job.ErrorMessage)
				}
			}
		}
	})
}
