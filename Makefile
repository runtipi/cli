# Makefile for runtipi CLI

# Build values
VERSION := nightly
COMMIT := $(shell git rev-parse HEAD)
DATE := $(shell date +%Y-%m-%dT%H:%M:%S%z)

# Define the root folder path
ROOT_FOLDER_HOST := ~/temp/runtipi

# Set environment variables for golang target platform if these are passed via command line
ifdef GOOS
ifdef GOARCH
	export GOOS
	export GOARCH
endif
endif

# Read golang target platform from the system
TARGET_OS := $$(go env GOOS)
TARGET_ARCH := $$(go env GOARCH)

# Main target that creates the directory and runs the program
.PHONY: run
run: current_platform
	@mkdir -p $(ROOT_FOLDER_HOST)
	@ROOT_FOLDER_HOST=$(ROOT_FOLDER_HOST) go run -ldflags="-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildDate=$(DATE)" cmd/runtipi/main.go $(ARGS)

# Build the program
.PHONY: build
build: current_platform
	@go build -o runtipi-cli-$(TARGET_OS)-$(TARGET_ARCH) cmd/runtipi/main.go

# Clean build artifacts
.PHONY: clean
clean:
	@rm -f runtipi-cli-*

# List all available golang target platforms
.PHONY: available_platforms
available_platforms:
	@go tool dist list

# Print which golang target platform is currently set
.PHONY: current_platform
current_platform:
	@echo "Current golang target platform: $(TARGET_OS)/$(TARGET_ARCH)"

# String formats to use for output of help command
header_line_format='--------------------------------------\n%s\n--------------------------------------\n'
line_format='  %-40s - %s\n' # Fix padding of first column to 40 chars

# Help command
.PHONY: help
help:
	@printf -- $(header_line_format) "Available commands"
	@printf $(line_format) "make run ARGS=\"command arguments\"" 	"Run the program with specified arguments"
	@printf $(line_format) "make build" 							"Build the program"
	@printf $(line_format) "make clean"								"Clean build artifacts"
	@printf $(line_format) "make available_platforms"				"List golang target platforms available for the build"
	@printf $(line_format) "make current_platform"					"Print current golang target platform"
	@printf $(line_format) "make help"								"Show this help message"
	@printf "\n"
	@printf -- $(header_line_format) "Example usage"
	@printf $(line_format) "make run ARGS=\"start\""				"Run the program with 'start' command"
	@printf $(line_format) "make run ARGS=\"update nightly\""		"Run the program with 'update nightly' command"
	@printf $(line_format) "make build GOOS=darwin GOARCH=arm64"	"Build the program for golang target platform 'darwin/arm64' (macOS silicon)"
