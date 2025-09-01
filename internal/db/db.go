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
		`INSERT INTO jobs (id, type, payload, status, created_at, updated_at)
         VALUES ($1, $2, $3, $4, $5, $6)`,
		job.ID, job.Type, job.Payload, job.Status, time.Now(), time.Now(),
	)
	return err
}

func (d *DB) GetJob(ctx context.Context, id string) (models.Job, error) {
	var job models.Job
	row := d.pool.QueryRow(ctx,
		`SELECT id, type, payload, status, created_at, updated_at FROM jobs WHERE id = $1`, id,
	)
	err := row.Scan(&job.ID, &job.Type, &job.Payload, &job.Status, &job.CreatedAt, &job.UpdatedAt)
	return job, err
}

func (d *DB) UpdateJobStatus(ctx context.Context, id string, status models.JobStatus) error {
	_, err := d.pool.Exec(ctx,
		`UPDATE jobs SET status = $1, updated_at = $2 WHERE id = $3`,
		status, time.Now(), id,
	)
	return err
}
