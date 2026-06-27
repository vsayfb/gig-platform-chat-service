.PHONY: run build tidy

run:
	go run ./cmd/api

build:
	go build -o bin/chat-service ./cmd/api

tidy:
	go mod tidy
