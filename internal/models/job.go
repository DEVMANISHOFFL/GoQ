package models

import "time"

type JobStatus string

const (
	StatusPending JobStatus = "pending"
	StatusRunning JobStatus = "running"
	StatusSuccess JobStatus = "success"
	StatusFailed  JobStatus = "failed"
)

type Job struct {
	ID           string
	Type         string
	Payload      string
	Status       JobStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Retries      int
	MaxRetries   int
	ErrorMessage string
}
