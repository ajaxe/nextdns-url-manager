# Makefile for NextDNS Client project
# Delegates build operations to build.ps1

SHELL := pwsh

# Default target
.PHONY: all
all: build

# Build the application
.PHONY: build
build: clean
	@pwsh -NoProfile -File "./build.ps1" -Task build

# Build cross-platform binaries
.PHONY: dist
dist: clean
	@pwsh -NoProfile -File "./build.ps1" -Task dist

# Windows builds
.PHONY: win64
win64:
	@pwsh -NoProfile -File "./build.ps1" -Task win64

.PHONY: win32
win32:
	@pwsh -NoProfile -File "./build.ps1" -Task win32

# Linux builds
.PHONY: linux64
linux64:
	@pwsh -NoProfile -File "./build.ps1" -Task linux64

.PHONY: linux32
linux32:
	@pwsh -NoProfile -File "./build.ps1" -Task linux32

# macOS builds
.PHONY: mac64
mac64:
	@pwsh -NoProfile -File "./build.ps1" -Task mac64

.PHONY: macarm64
macarm64:
	@pwsh -NoProfile -File "./build.ps1" -Task macarm64

# All cross-platform binaries
.PHONY: cross-platform
cross-platform:
	@pwsh -NoProfile -File "./build.ps1" -Task cross-platform

# Clean build artifacts
.PHONY: clean
clean:
	@pwsh -NoProfile -File "./build.ps1" -Task clean

# Run the application
.PHONY: run
run:
	@go run ./cmd/main.go

# Test the application
.PHONY: test
test:
	@go test ./...

# Lint the code
.PHONY: lint
lint:
	@go fmt ./...
	@go vet ./...

# Help
.PHONY: help
help:
	@echo "NextDNS Client Makefile"
	@echo "======================"
	@echo "make build        - Build the application"
	@echo "make dist         - Create distribution directory"
	@echo "make cross-platform - Build all cross-platform binaries"
	@echo "make win64        - Build Windows x64 binary"
	@echo "make win32        - Build Windows x32 binary"
	@echo "make linux64      - Build Linux x64 binary"
	@echo "make linux32      - Build Linux x32 binary"
	@echo "make mac64        - Build macOS x64 binary"
	@echo "make macarm64     - Build macOS ARM64 binary"
	@echo "make run          - Run the application"
	@echo "make test         - Run tests"
	@echo "make lint         - Format and vet code"
	@echo "make clean        - Clean build artifacts"
	@echo "make help         - Show this help"

# Default target
.DEFAULT_GOAL := build
