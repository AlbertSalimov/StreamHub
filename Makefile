.PHONY: run test build vet fmt check up down logs db-ready

run:
	set -a; . ./.env; set +a; go run ./cmd/api

test:
	go test ./...

build:
	go build -o bin/streamhub ./cmd/api

vet:
	go vet ./...

fmt:
	gofmt -w cmd internal

check: fmt vet test build

up:
	docker compose up -d postgres

down:
	docker compose down

logs:
	docker compose logs -f postgres

db-ready:
	docker compose exec postgres pg_isready -U streamhub -d streamhub