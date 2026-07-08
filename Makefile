.PHONY: build run test test-race lint migrate-up migrate-down db-up db-down db-reset tidy help

BIN := bin/server
CMD := ./cmd/server
MIGRATIONS_DIR := db/migrations

ifneq (,$(wildcard .env))
include .env
export
endif

build:
	go build -o $(BIN) $(CMD)

run:
	go run $(CMD)

test:
	go test ./...

test-race:
	go test -race ./...

lint:
	golangci-lint run ./...

migrate-up:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" down

db-up:
	docker compose up -d

db-down:
	docker compose down

db-reset:
	docker compose down -v

tidy:
	go mod tidy

help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //'