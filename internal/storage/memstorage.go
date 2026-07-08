package storage

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

type MemStorage struct {
	checks map[uuid.UUID]Task
	mtx    sync.RWMutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		checks: make(map[uuid.UUID]Task),
	}
}

func (s *MemStorage) CreateTask(_ context.Context, id uuid.UUID) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.checks[id] = Task{Status: Pending}
	return nil
}

func (s *MemStorage) CompleteTask(_ context.Context, id uuid.UUID, results []Result) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.checks[id] = Task{Status: Done, Result: results}
	return nil
}

func (s *MemStorage) GetResult(_ context.Context, id uuid.UUID) (Task, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	if _, ok := s.checks[id]; !ok {
		return Task{}, ErrTaskDoesNotExist
	}
	return s.checks[id], nil

}

func (s *MemStorage) Close() {

}
