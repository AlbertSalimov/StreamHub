.PHONY: run build vet fmt db-ready test test-integration check docker-build docker-up docker-down docker-logs docker-ps

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

db-ready:
	docker compose exec postgres pg_isready -U streamhub -d streamhub

docker-build:
	docker compose build

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f api

docker-ps:
	docker compose ps