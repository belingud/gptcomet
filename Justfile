# Default goal: help
help:
    @just --list

# Install the environment
install:
    @echo "🚀 Creating virtual environment using pyenv and PDM"
    pdm install

# Run code quality tools
check:
    @echo "🚀 Checking pdm lock file consistency with 'pyproject.toml': Running pdm lock --check"
    pdm lock --check
    @echo "🚀 Linting code: Running pre-commit"
    pdm run pre-commit run -a
    @echo "🚀 Linting with ruff"
    pdm run ruff check . --config pyproject.toml --exclude tests
    @echo "🚀 Checking for obsolete dependencies: Running deptry"
    pdm run deptry .

# Format code with ruff and isort
format:
    @echo "🚀 Formatting code: Running ruff"
    pdm run ruff format . --config pyproject.toml
    @echo "🚀 Formatting code: Running isort"
    pdm run isort . --settings-path pyproject.toml
    @echo "🚀 Formatting code: Running go fmt"
    go fmt ./...

# Test the code with pytest
test:
    @echo "🚀 Testing code: Running pytest"
    pdm run pytest --cov --cov-config=pyproject.toml --cov-report=xml tests

# Clean build artifacts
clean-build:
    rm -rf dist

# Build wheel file
build: clean-build
    @echo "🚀 Creating wheel file"
    pdm build

# Publish a release to PyPI
publish:
    @echo "🚀 Publishing."
    rm -rf dist
    uv-publish

# Publish a release to TestPyPI
publish-test:
    @echo "🚀 Publishing to testpypi."
    pdm publish -r testpypi --username __token__

# Build and publish
build-and-publish: build publish

# Update changelog
changelog:
    git cliff -l --prepend CHANGELOG.md
