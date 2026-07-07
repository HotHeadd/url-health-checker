package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
	"url-checker/internal/storage"

	"github.com/google/uuid"
)

type CheckRequest struct {
	URLs []string `json:"urls"`
}

type CheckResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type HealthResponse struct {
	Status string `json:"status"`
}
type CheckIdResp struct {
	ID      string           `json:"id"`
	Status  string           `json:"status"`
	Results []storage.Result `json:"results,omitempty"`
}

func (s *Server) HandleGetHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := HealthResponse{Status: "ok"}
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		s.logger.Error("failed to encode response", "error", err)
	}
}

func (s *Server) HandlePostChecks(w http.ResponseWriter, r *http.Request) {
	req := CheckRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		s.logger.Warn("invalid request body", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	id := uuid.New()
	err = s.storage.CreateTask(r.Context(), id)
	if err != nil {
		s.logger.Error("failed to create task in db", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	s.checkGroup.Go(func() {
		s.logger.Info("check started", "id", id, "urls", len(req.URLs))
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		start := time.Now()
		result := s.checker.CheckAll(ctx, req.URLs)
		end := time.Since(start).Milliseconds()
		err = s.storage.CompleteTask(ctx, id, result)
		if err != nil {
			s.logger.Error("failed to complete task in db", "error", err)
		} else {
			s.logger.Info("check completed", "id", id, "duration_ms", end)
		}
	})

	resp := CheckResponse{
		ID:     id.String(),
		Status: string(storage.Pending),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		s.logger.Error("failed to encode response", "error", err)
	}
}

func (s *Server) HandleGetCheckId(w http.ResponseWriter, r *http.Request) {
	idRaw := r.PathValue("id")
	id, err := uuid.Parse(idRaw)
	if err != nil {
		s.logger.Warn("invalid id format in request", "id", idRaw, "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	task, err := s.storage.GetResult(r.Context(), id)
	if errors.Is(err, storage.ErrTaskDoesNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		s.logger.Error("error getting result from db", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := CheckIdResp{
		ID:      idRaw,
		Status:  string(task.Status),
		Results: task.Result,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		s.logger.Error("failed to encode response", "error", err)
	}
}
