package api

import (
	"net/http"
	"url-checker/internal/checker"
	"url-checker/internal/storage"
)

type Server struct {
	mux     *http.ServeMux
	checker *checker.Checker
	storage *storage.Storage
}

func NewServer() *Server {
	s := &Server{
		mux:     http.NewServeMux(),
		checker: checker.NewChecker(),
		storage: storage.NewStorage(),
	}
	s.mux.HandleFunc("GET /health", s.HandleGetHealth)
	s.mux.HandleFunc("POST /checks", s.HandlePostChecks)
	s.mux.HandleFunc("GET /checks/{id}", s.HandleGetCheckId)
	return s
}

func (s *Server) Routes() *http.ServeMux {
	return s.mux
}
