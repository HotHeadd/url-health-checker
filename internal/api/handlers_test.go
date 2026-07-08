package api

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"url-checker/internal/storage"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestServer() (*Server, storage.Storage) {
	st := storage.NewMemStorage()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	return NewServer(logger, st), st
}

func TestHealthHandler(t *testing.T) {
	s, _ := createTestServer()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	s.Routes().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resDecoded HealthResponse
	err := json.NewDecoder(rec.Body).Decode(&resDecoded)
	assert.NoError(t, err)
	assert.Equal(t, "ok", resDecoded.Status)
}

func TestPostHandlerBroken(t *testing.T) {
	s, _ := createTestServer()
	body := strings.NewReader(`{broken}`)
	req := httptest.NewRequest(http.MethodPost, "/checks", body)
	rec := httptest.NewRecorder()
	s.Routes().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPostHandlerCorrect(t *testing.T) {
	s, _ := createTestServer()
	body := strings.NewReader(`{"urls":[]}`)
	req := httptest.NewRequest(http.MethodPost, "/checks", body)
	rec := httptest.NewRecorder()
	s.Routes().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusAccepted, rec.Code)
	var resDecoded CheckResponse
	err := json.NewDecoder(rec.Body).Decode(&resDecoded)
	assert.NoError(t, err)
	assert.Equal(t, "pending", resDecoded.Status)
	_, err = uuid.Parse(resDecoded.ID)
	assert.NoError(t, err)
}

func TestCheckIdDontExist(t *testing.T) {
	s, _ := createTestServer()
	missingID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/checks/"+missingID.String(), nil)
	rec := httptest.NewRecorder()
	s.Routes().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestCheckWrongId(t *testing.T) {
	s, _ := createTestServer()
	req := httptest.NewRequest(http.MethodGet, "/checks/wrongID", nil)
	rec := httptest.NewRecorder()
	s.Routes().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCheckIdCorrect(t *testing.T) {
	s, strg := createTestServer()
	id := uuid.New()
	ctx := context.Background()
	results := []storage.Result{
		{
			URL:        "google.com",
			StatusCode: http.StatusOK,
			Duration:   3,
			Err:        "",
		},
	}
	require.NoError(t, strg.CreateTask(ctx, id))
	require.NoError(t, strg.CompleteTask(ctx, id, results))

	req := httptest.NewRequest(http.MethodGet, "/checks/"+id.String(), nil)
	rec := httptest.NewRecorder()
	s.Routes().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resDecoded CheckIDResp
	err := json.NewDecoder(rec.Body).Decode(&resDecoded)
	assert.NoError(t, err)
	assert.Equal(t, id.String(), resDecoded.ID)
	assert.Equal(t, string(storage.Done), resDecoded.Status)
	assert.Equal(t, results, resDecoded.Results)
}
