.PHONY: build test lint clean install dev snapshot release help

# Variables
BINARY_NAME := rtmx
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-s -w -X github.com/rtmx-ai/rtmx-go/internal/cmd.Version=$(VERSION) -X github.com/rtmx-ai/rtmx-go/internal/cmd.Commit=$(COMMIT) -X github.com/rtmx-ai/rtmx-go/internal/cmd.Date=$(DATE)"

# Default target
all: build

## build: Build the binary
build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/rtmx

## dev: Build with race detector for development
dev:
	go build -race $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/rtmx

## install: Install to $GOPATH/bin
install:
	go install $(LDFLAGS) ./cmd/rtmx

## test: Run tests
test:
	go test -v -race -coverprofile=coverage.out ./...

## test-short: Run short tests only
test-short:
	go test -v -short ./...

## coverage: Show test coverage in browser
coverage: test
	go tool cover -html=coverage.out

## lint: Run linter
lint:
	golangci-lint run

## fmt: Format code
fmt:
	go fmt ./...
	goimports -w .

## tidy: Tidy and verify dependencies
tidy:
	go mod tidy
	go mod verify

## clean: Remove build artifacts
clean:
	rm -rf bin/ dist/ coverage.out

## snapshot: Build snapshot release (local testing)
snapshot:
	goreleaser release --snapshot --clean

## release-check: Validate release configuration
release-check:
	goreleaser check

## build-all: Build for all platforms
build-all:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 ./cmd/rtmx
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-arm64 ./cmd/rtmx
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 ./cmd/rtmx
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 ./cmd/rtmx
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe ./cmd/rtmx

## parity: Run parity tests against Python CLI
parity:
	@echo "Running parity tests..."
	go test -v -tags=parity ./test/parity/...

## help: Show this help
help:
	@echo "RTMX Go CLI - Makefile targets"
	@echo ""
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/  /'
