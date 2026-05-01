.PHONY: help build run test lint clean fmt protolint golint up down logs docker-clean

lint: golint

help:
	@echo "Available commands:"
	@echo "  make build"
	@echo "  make run"
	@echo "  make test"
	@echo "  make lint"
	@echo "  make fmt"
	@echo "  make up"
	@echo "  make down"
	@echo "  make logs"
	@echo "  make clean"

build:
	go build -o bin/api ./api

run:
	go run ./api -config api/config.yaml

test:
	go test -v -cover ./...

golint:
	/home/strees111/go/bin/golangci-lint run -E gocritic -v ./...

fmt:
	go fmt ./...
	goimports -w .

up:
	docker compose up --build -d

down:
	docker compose down

logs:
	docker compose logs -f go_service

clean:
	rm -rf bin/
	go clean

docker-clean:
	docker compose down -v