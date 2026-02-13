.PHONY: run build test sqlc-generate api-generate fmt

run:
	go run ./cmd/tui

build:
	@mkdir -p bin
	go build -o bin/tui ./cmd/tui

test:
	go test ./...

sqlc-generate:
	GOARCH=amd64 GOOS=linux go run ./scripts/generate_sqlc

api-generate:
	GOARCH=amd64 GOOS=linux go run ./scripts/generate_openapi_client

fmt:
	go fmt ./...
