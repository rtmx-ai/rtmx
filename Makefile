.PHONY: build test lint clean install dev snapshot release help

# Variables
BINARY_NAME := rtmx
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-s -w -X github.com/rtmx-ai/rtmx/internal/cmd.Version=$(VERSION) -X github.com/rtmx-ai/rtmx/internal/cmd.Commit=$(COMMIT) -X github.com/rtmx-ai/rtmx/internal/cmd.Date=$(DATE)"

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

## lint: Run linter (golangci-lint v2)
lint:
	@command -v golangci-lint >/dev/null 2>&1 || $(HOME)/go/bin/golangci-lint version >/dev/null 2>&1 || \
		{ echo "golangci-lint not found. Install: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b \$$(go env GOPATH)/bin"; exit 1; }
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run ./... || $(HOME)/go/bin/golangci-lint run ./...

## hooks: Install pre-commit hooks
hooks:
	git config core.hooksPath .githooks
	@echo "Pre-commit hooks installed from .githooks/"

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

## ci: Run local CI matching GitHub CI (build, test, coverage, lint, vet, markers)
ci:
	@echo "=== Build ===" && go build -v ./... && echo "PASS Build"
	@echo "=== Test + Coverage ===" && go test -race -coverprofile=coverage.out -covermode=atomic ./... && echo "PASS Test"
	@echo "=== Coverage Threshold ===" && \
		COVERAGE=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | tr -d '%') && \
		echo "Total coverage: $${COVERAGE}%" && \
		if [ $$(echo "$${COVERAGE} < 70" | bc -l) -eq 1 ]; then echo "FAIL Coverage below 70%"; exit 1; fi && \
		echo "PASS Coverage"
	@echo "=== Vet ===" && go vet ./... && echo "PASS Vet"
	@echo "=== Marker Compliance ===" && \
		TOTAL=$$(grep -r 'func Test' internal/ test/ pkg/ --include='*_test.go' -l | wc -l | tr -d ' ') && \
		MARKED=$$(grep -r 'rtmx\.Req(t,' internal/ test/ pkg/ --include='*_test.go' -l | wc -l | tr -d ' ') && \
		PCT=$$((MARKED * 100 / TOTAL)) && \
		echo "Markers: $${MARKED}/$${TOTAL} ($${PCT}%)" && \
		if [ "$$PCT" -lt 80 ]; then echo "FAIL Marker compliance below 80%"; exit 1; fi && \
		echo "PASS Markers"
	@echo "=== All CI checks passed ==="
	@rm -f coverage.out
