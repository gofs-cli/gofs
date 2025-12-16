$(shell cp -n .env.example .env)
include .env
export

run: dbup
	@go tool air
.PHONY: run

build:
	@go build -o bin/app ./cmd/server/main.go
.PHONY: build

lint:
	@golangci-lint run
.PHONY: lint

vuln:
	@go tool govulncheck ./...
.PHONY: vuln

test:
	@go test -v ./...
.PHONY: test

dbup:
	@if [ -n "$$DSN" ]; then \
		docker compose -f docker/docker-compose.yml up -d postgres; \
    fi
.PHONY: dbup

dbdown:
	@docker compose -f docker/docker-compose.yml down
.PHONY: dbdown

