.PHONY: run build vet fmt up down logs db-ready test test-integration check

run:
	set -a; . ./.env; set +a; go run ./cmd/api

test:
	go test ./...

test-integration:
	go test ./internal/content -run TestRepositoryIntegration -v

build:
	go build -o bin/streamhub ./cmd/api

vet:
	go vet ./...

fmt:
	gofmt -w cmd internal

check:
	gofmt -w cmd internal
	go vet ./...
	go test ./...
	go build ./cmd/api

up:
	docker compose up -d postgres

down:
	docker compose down

logs:
	docker compose logs -f postgres

db-ready:
	docker compose exec postgres pg_isready -U streamhub -d streamhub