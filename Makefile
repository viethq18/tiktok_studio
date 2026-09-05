.DEFAULT_GOAL := help
SHELL := /bin/bash

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

setup: ## Create .env and install frontend dependencies
	@test -f .env || (cp .env.example .env && echo "created .env")
	cd frontend && npm install

infra: ## Start Postgres, Redis and MinIO
	docker compose up -d
	@echo "postgres :5432  redis :6379  minio :9100 (console :9101)"

infra-down: ## Stop infrastructure
	docker compose down

api: ## Run the HTTP API (migrations run at boot)
	cd backend && go run ./cmd/server

worker: ## Run the generation/export worker
	cd backend && go run ./cmd/worker

web: ## Run the Next.js app
	cd frontend && npm run dev

mobile: ## Run the Flutter app (needs a simulator or device)
	cd mobile && flutter run

build: ## Compile both Go binaries and the Next.js bundle
	cd backend && go build -o bin/server ./cmd/server && go build -o bin/worker ./cmd/worker
	cd frontend && npm run build

test: ## Run the Go test suite
	cd backend && go test ./...

test-mobile: ## Run the Flutter test suite
	cd mobile && flutter test

test-mobile-device: ## Run the on-device tests (needs a simulator and a running API)
	cd mobile && flutter test integration_test

apigen: ## Regenerate the OpenAPI document and the Dart/TypeScript clients
	cd backend && go run ./cmd/apigen ..

golden: ## Rewrite the export renderer's golden slide
	cd backend && go test ./internal/export -update-golden

smoke: ## Drive the whole MVP flow against a running stack
	./scripts/smoke.sh

e2e: ## Browser tests for the editor (needs API + worker + web running)
	cd frontend && npx playwright test

e2e-install: ## One-time: download the browser Playwright drives
	cd frontend && npx playwright install chromium

fmt: ## Format Go sources
	cd backend && go fmt ./...

reset-db: ## Drop and recreate the database (destructive)
	docker compose exec -T postgres psql -U tks -d postgres -c "DROP DATABASE IF EXISTS tks;" -c "CREATE DATABASE tks;"

.PHONY: help setup infra infra-down api worker web mobile build test golden apigen smoke e2e e2e-install fmt reset-db
