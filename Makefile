BIN := "./bin/migratory"
DOCKER_IMG = "migratory:develop"

# ! PUBLIC SCHEMA WILL BE DROPPED DURING INTEGRATION TESTS ! DO NOT CHANGE DATA SOURCE OR USE ONLY TEST DATABASE
DB_NAME = test
DB_USER = postgres
DB_PASSWORD = password
POSTGRES_PORT = 5432
POSTGRES_DSN = postgresql://$(DB_USER):$(DB_PASSWORD)@localhost:$(POSTGRES_PORT)/$(DB_NAME)?sslmode=disable

COMPOSE_PATH = ./deployments/docker-compose.yml
COMPOSE_ENV = DB_NAME=$(DB_NAME) DB_USER=$(DB_USER) DB_PASSWORD=$(DB_PASSWORD) POSTGRES_PORT=$(POSTGRES_PORT)

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/migratory/main.go

run: build
	$(BIN)

lint-install:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.55.2

lint: lint-install
	golangci-lint run ./... -v

test:
	go test -race ./internal/...

integration-test: postgres-up
	go test ./test/ -dsn $(POSTGRES_DSN) -tags integration

postgres-up:
	$(COMPOSE_ENV) docker compose -f $(COMPOSE_PATH) up -d

postgres-down:
	$(COMPOSE_ENV) docker compose -f $(COMPOSE_PATH) down

.PHONY: build run lint-install lint test integration-test postgres-up postgres-down
