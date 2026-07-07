package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"
	"url-checker/internal/checker"
	"url-checker/internal/storage"
)

type Server struct {
	mux        *http.ServeMux
	checker    *checker.Checker
	storage    storage.Storage
	logger     *slog.Logger
	checkGroup *sync.WaitGroup
}

func NewServer(logger *slog.Logger) (*Server, error) {
	dtUrl := os.Getenv("DATABASE_URL")
	dtCtx := context.Background()
	storage, err := storage.NewPgStorage(dtCtx, dtUrl)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	logger.Info("database is up")
	// storage := storage.NewMemStorage()
	s := &Server{
		mux:        http.NewServeMux(),
		checker:    checker.NewChecker(),
		storage:    storage,
		logger:     logger,
		checkGroup: &sync.WaitGroup{},
	}
	s.mux.HandleFunc("GET /health", s.HandleGetHealth)
	s.mux.HandleFunc("POST /checks", s.HandlePostChecks)
	s.mux.HandleFunc("GET /checks/{id}", s.HandleGetCheckId)
	return s, nil
}

func (s *Server) Routes() *http.ServeMux {
	return s.mux
}

func (s *Server) WaitChecks() <-chan struct{} {
	done := make(chan struct{})
	go func() {
		s.checkGroup.Wait()
		close(done)
	}()

	return done
}

func (s *Server) Close() {
	done := s.WaitChecks() // дожидаемся активных проверок
	select {
	case <-done:
	case <-time.After(15 * time.Second):
	}
	s.storage.Close()
}
