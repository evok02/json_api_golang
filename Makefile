.DEFAULT_GOAL := run
.PHONY: build run test

fmt:
	@go fmt ./...

vet: fmt
	@go vet ./...

build:
	@go build -o bin/gobank

run: build
	@go run .

test:
	@go test -v ./...
