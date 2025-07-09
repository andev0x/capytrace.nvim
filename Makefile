PLUGIN_NAME = capytrace
GO_BINARY = bin/$(PLUGIN_NAME)
GO_SOURCE = main.go
GO_PACKAGES = recorder exporter

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
	go test ./recorder ./exporter

dev: build
	@echo "Development build complete"

# Go module initialization
go-mod-init:
	go mod init github.com/andev0x/capytrace.nvim
	go mod tidy

# Format Go code
fmt:
	go fmt ./recorder/*.go
	go fmt ./exporter/*.go
	go fmt ./main.go

# Check for Go dependencies
check-deps:
	@which go > /dev/null || (echo "Go is not installed. Please install Go first." && exit 1)
	@echo "Go version: $$(go version)"