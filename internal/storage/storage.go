package storage

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Result struct {
	URL        string        `json:"url"`
	StatusCode int           `json:"status_code"`
	Duration   time.Duration `json:"request_duration"`
	Err        string        `json:"error"`
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

type Storage struct {
	checks map[uuid.UUID]Task
	mtx    sync.RWMutex
}

var ErrTaskDoesNotExist error = errors.New("task not found")

func NewStorage() *Storage {
	return &Storage{
		checks: make(map[uuid.UUID]Task),
	}
}

func (s *Storage) SetResult(id uuid.UUID, res Task) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.checks[id] = res
}

func (s *Storage) GetResult(id uuid.UUID) (res Task, err error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	if _, ok := s.checks[id]; !ok {
		return Task{}, ErrTaskDoesNotExist
	}
	return s.checks[id], nil

}
