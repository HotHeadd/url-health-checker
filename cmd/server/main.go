// Package main executes program.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"url-checker/internal/api"
	"url-checker/internal/storage"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	logger := slog.New(slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))

	dtURL := os.Getenv("DATABASE_URL")
	dtCtx := context.Background()
	st, err := storage.NewPgStorage(dtCtx, dtURL)
	if err != nil {
		logger.Error("server creation failed", "error", err)
		return
	}
	logger.Info("database is up")

	s := api.NewServer(logger, st)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      s.Routes(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
		}
	}()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil { // говорим серверу чтобы не принимал новые соединения
		logger.Error("graceful shutdown failed", "error", err)
	}
	s.Close()
}
