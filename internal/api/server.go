package api

import (
	"log/slog"
	"net/http"
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

func NewServer(logger *slog.Logger, st storage.Storage) *Server {
	s := &Server{
		mux:        http.NewServeMux(),
		checker:    checker.NewChecker(),
		storage:    st,
		logger:     logger,
		checkGroup: &sync.WaitGroup{},
	}
	s.mux.HandleFunc("GET /health", s.HandleGetHealth)
	s.mux.HandleFunc("POST /checks", s.HandlePostChecks)
	s.mux.HandleFunc("GET /checks/{id}", s.HandleGetCheckID)
	return s
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
