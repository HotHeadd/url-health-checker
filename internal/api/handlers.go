package api

import (
	"context"
	"encoding/json"
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
	json.NewEncoder(w).Encode(response) // TBD: log error
}

func (s *Server) HandlePostChecks(w http.ResponseWriter, r *http.Request) {
	req := CheckRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // TBD: log and describe
		return
	}
	id := uuid.New()
	s.storage.SetResult(id, storage.Task{
		Status: storage.Pending,
	})
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		result := s.checker.CheckAll(ctx, req.URLs)
		s.storage.SetResult(id, storage.Task{
			Status: storage.Done,
			Result: result,
		})

	}()

	resp := CheckResponse{
		ID:     id.String(),
		Status: string(storage.Pending),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp) // TBD: log error
}

func (s *Server) HandleGetCheckId(w http.ResponseWriter, r *http.Request) {
	idRaw := r.PathValue("id")
	id, err := uuid.Parse(idRaw)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // TBD: log and describe (bad id format)
		return
	}
	task, err := s.storage.GetResult(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound) // TBD: log and describe (task does not exist)
		return
	}

	resp := CheckIdResp{
		ID:      idRaw,
		Status:  string(task.Status),
		Results: task.Result,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp) // TBD: log error
}
