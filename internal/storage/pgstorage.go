package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgStorage struct {
	pool *pgxpool.Pool
}

var ErrDBUrlNotSet error = errors.New("database url not set in env")

func NewPgStorage(ctx context.Context, connString string) (*PgStorage, error) {
	if connString == "" {
		return nil, ErrDBUrlNotSet
	}
	pool, err := pgxpool.New(
		ctx,
		connString,
	)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}
	return &PgStorage{pool}, nil
}

const (
	createTaskQuery = "INSERT INTO tasks(id, proc_status) VALUES($1, $2)"
	updateTaskQuery = "UPDATE tasks SET proc_status='done' WHERE id=$1"
	addResultQuery  = "INSERT INTO results(task_id, url, status_code, duration_ms, error) VALUES($1, $2, $3, $4, $5)"
	getTaskQuery    = "SELECT proc_status FROM tasks WHERE id = $1"
	getResultsQuery = "SELECT url, status_code, duration_ms, error FROM results WHERE task_id = $1 ORDER BY id"
)

func (s *PgStorage) CreateTask(ctx context.Context, id uuid.UUID) error {
	_, err := s.pool.Exec(ctx, createTaskQuery, id, Pending)
	return err
}

func (s *PgStorage) CompleteTask(ctx context.Context, id uuid.UUID, results []Result) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, updateTaskQuery, id)
	if err != nil {
		return err
	}
	for _, result := range results {
		_, err := tx.Exec(ctx, addResultQuery, id, result.URL, result.StatusCode, result.Duration, result.Err)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (s *PgStorage) GetResult(ctx context.Context, id uuid.UUID) (Task, error) {
	var status Status
	err := s.pool.QueryRow(ctx, getTaskQuery, id).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return Task{}, ErrTaskDoesNotExist
	}
	if err != nil {
		return Task{}, fmt.Errorf("get task: %w", err)
	}

	task := Task{Status: status}

	rows, err := s.pool.Query(ctx, getResultsQuery, id)
	if err != nil {
		return Task{}, fmt.Errorf("get task: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var r Result
		err = rows.Scan(&r.URL, &r.StatusCode, &r.Duration, &r.Err)
		if err != nil {
			return Task{}, fmt.Errorf("get task: %w", err)
		}
		task.Result = append(task.Result, r)
	}
	if err = rows.Err(); err != nil {
		return Task{}, fmt.Errorf("get task: %w", err)
	}
	return task, nil
}

func (s *PgStorage) Close() {
	s.pool.Close()
}
