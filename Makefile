PLUGIN_NAME = debugstory
GO_BINARY = bin/$(PLUGIN_NAME)
GO_SOURCE = go/main.go
GO_PACKAGES = go/recorder go/exporter

.PHONY: all build clean install test

all: build

build:
	@echo "Building $(PLUGIN_NAME)..."
	@mkdir -p bin
	@cd go && go build -o ../$(GO_BINARY) .

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/

install: build
	@echo "Plugin ready for installation"

test:
	@echo "Running Go tests..."
	@cd go && go test ./...

dev: build
	@echo "Development build complete"

# Go module initialization
go-mod-init:
	@cd go && go mod init github.com/debugstory
	@cd go && go mod tidy

# Format Go code
fmt:
	@cd go && go fmt ./...

# Check for Go dependencies
check-deps:
	@which go > /dev/null || (echo "Go is not installed. Please install Go first." && exit 1)
	@echo "Go version: $$(go version)"