package storage

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type Result struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code"`
	Duration   int64  `json:"request_duration_ms"`
	Err        string `json:"error"`
}

type Status string

const (
	Done    Status = "done"
	Pending Status = "pending"
	Failed  Status = "failed"
)

type Task struct {
	Status Status
	Result []Result
}

type Storage interface {
	CreateTask(ctx context.Context, id uuid.UUID) error
	CompleteTask(ctx context.Context, id uuid.UUID, results []Result) error
	GetResult(ctx context.Context, id uuid.UUID) (Task, error)
	Close()
}

var ErrTaskDoesNotExist error = errors.New("task not found")
