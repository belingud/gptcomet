# Default goal: help
help:
    @just --list

# ==============
# Go Commands
# ==============

# Run Go vet and staticcheck
check:
    @echo "🚀 Running Go vet and staticcheck"
    go vet ./...
    staticcheck ./...

# Format Go code
format:
    @echo "🚀 Formatting Go code"
    go fmt ./...
    goimports -w .

# Run Go tests with coverage
test:
    @echo "🚀 Running Go tests"
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html

# Build Go binaries
build:
    @echo "🚀 Building Go binaries"
    go build -o bin/gptcomet ./cmd

# Cross-compile Go binaries
build-all:
    @echo "🚀 Cross-compiling Go binaries"
    goreleaser build --snapshot --rm-dist

# Clean Go build artifacts
clean:
    @echo "🚀 Cleaning Go build artifacts"
    rm -rf bin/ dist/ coverage.out coverage.html

# ==============
# Python Commands
# ==============

# Install Python dependencies
install:
    @echo "🚀 Installing Python dependencies"
    uv sync

# Run Python tests
test-py:
    @echo "🚀 Running Python tests"
    uv run pytest --cov --cov-config=pyproject.toml --cov-report=xml tests

# Build Python wheel
build-py:
    @echo "🚀 Building Python wheel"
    uv build

# Publish Python package
publish-py:
    @echo "🚀 Publishing Python package"
    uv publish

# ==============
# Release Commands
# ==============

# Create a new release
release:
    @echo "🚀 Creating new release"
    goreleaser release --rm-dist
    uv publish

# ==============
# Utility Commands
# ==============

# Update changelog
changelog:
    git cliff -l --prepend CHANGELOG.md

# Generate coverage report
coverage:
    @echo "🚀 Generating coverage report"
    go tool cover -html=coverage.out -o coverage.html
    open coverage.html
