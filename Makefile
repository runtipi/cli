# Makefile for runtipi CLI

# Build values
VERSION := nightly
COMMIT := $(shell git rev-parse HEAD)
DATE := $(shell date +%Y-%m-%dT%H:%M:%S%z)

# Define the root folder path
ROOT_FOLDER_HOST := ~/temp/runtipi

# Main target that creates the directory and runs the program
.PHONY: run
run:
	@mkdir -p $(ROOT_FOLDER_HOST)
	@ROOT_FOLDER_HOST=$(ROOT_FOLDER_HOST) go run -ldflags="-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildDate=$(DATE)" cmd/runtipi/main.go $(ARGS)

# Build the program
.PHONY: build
build:
	@go build -o runtipi-cli cmd/runtipi/main.go

# Clean build artifacts
.PHONY: clean
clean:
	@rm -f runtipi-cli

# Help command
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make run ARGS=\"command arguments\"  - Run the program with specified arguments"
	@echo "  make build                          - Build the program"
	@echo "  make clean                          - Clean build artifacts"
	@echo "  make help                           - Show this help message"
	@echo ""
	@echo "Example usage:"
	@echo "  make run ARGS=\"start\"               - Run the program with 'start' command"
	@echo "  make run ARGS=\"update nightly\"       - Run the program with 'update nightly' command"
