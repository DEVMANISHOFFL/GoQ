package db

import (
	"context"
	"distributed-job-queue/internal/models"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	pool *pgxpool.Pool
}

func NewDB(conn string) (*DB, error) {
	pool, err := pgxpool.New(context.Background(), conn)
	if err != nil {
		return nil, err
	}
	return &DB{pool: pool}, nil
}

func (d *DB) SaveJob(ctx context.Context, job models.Job) error {
	_, err := d.pool.Exec(ctx,
		`INSERT INTO jobs (id, type, payload, status, created_at, updated_at, retries, max_retries, error_message)
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		job.ID, job.Type, job.Payload, job.Status, time.Now(), time.Now(),
		job.Retries, job.MaxRetries, job.ErrorMessage,
	)
	return err
}

func (d *DB) GetJob(ctx context.Context, id string) (models.Job, error) {
	var job models.Job
	row := d.pool.QueryRow(ctx,
		`SELECT id, type, payload, status, created_at, updated_at, retries, max_retries, error_message
         FROM jobs WHERE id = $1`, id,
	)
	err := row.Scan(&job.ID, &job.Type, &job.Payload, &job.Status,
		&job.CreatedAt, &job.UpdatedAt, &job.Retries, &job.MaxRetries, &job.ErrorMessage)
	return job, err
}

func (d *DB) UpdateJobStatus(ctx context.Context, id string, status models.JobStatus) error {
	_, err := d.pool.Exec(ctx,
		`UPDATE jobs SET status = $1, updated_at = $2 WHERE id = $3`,
		status, time.Now(), id,
	)
	return err
}

func (d *DB) UpdateJobFailure(ctx context.Context, id string, retries, maxRetries int, errMsg string) error {
	_, err := d.pool.Exec(ctx,
		`UPDATE jobs
         SET status = $1, retries = $2, max_retries = $3, error_message = $4, updated_at = $5
         WHERE id = $6`,
		models.StatusFailed, retries, maxRetries, errMsg, time.Now(), id,
	)
	return err
}

func (d *DB) UpdateJobRetry(ctx context.Context, id string, retries int, errMsg string) error {
	_, err := d.pool.Exec(ctx,
		`UPDATE jobs SET retries=$1, error_message=$2, updated_at=NOW() WHERE id=$3`,
		retries, errMsg, id,
	)
	return err
}
