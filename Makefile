.PHONY: run test build vet fmt check

run:
	go run ./cmd/api

test:
	go test ./...

build:
	go build -o bin/streamhub ./cmd/api

vet:
	go vet ./...

fmt:
	gofmt -w cmd internal

check: fmt vet test build