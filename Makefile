PLUGIN_NAME = capytrace
GO_BINARY = bin/$(PLUGIN_NAME)
GO_SOURCE = cmd/capytrace/main.go
GO_PACKAGES = internal/recorder internal/exporter internal/filter internal/models

.PHONY: all build clean install test

all: build

build:
	@echo "Building $(PLUGIN_NAME)..."
	@mkdir -p ./bin
	go build -o ./bin/$(PLUGIN_NAME) $(GO_SOURCE)

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/

install: build
	@echo "Plugin ready for installation"

test:
	@echo "Running Go tests..."
	go test ./internal/recorder ./internal/exporter ./internal/filter ./internal/models

dev: build
	@echo "Development build complete"

# Go module initialization
go-mod-init:
	go mod init github.com/andev0x/capytrace.nvim
	go mod tidy

# Format Go code
fmt:
	go fmt ./internal/recorder/*.go
	go fmt ./internal/exporter/*.go
	go fmt ./internal/filter/*.go
	go fmt ./internal/models/*.go
	go fmt ./cmd/capytrace/*.go

# Check for Go dependencies
check-deps:
	@which go > /dev/null || (echo "Go is not installed. Please install Go first." && exit 1)
	@echo "Go version: $$(go version)"