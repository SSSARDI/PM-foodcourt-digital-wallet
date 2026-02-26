.PHONY: run build tidy

run:
	go run ./cmd/api

build:
	go build -o bin/server ./cmd/api

tidy:
	go mod tidy

test:
	go test ./...
