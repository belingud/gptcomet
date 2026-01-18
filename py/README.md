# GPTComet Python Wrapper

## Overview

The Python wrapper is responsible for:
- Platform detection and binary file selection
- Process management and environment variable handling
- Easy distribution via PyPI

## Directory Structure

```
py/
├── gptcomet/
│   ├── __init__.py      # Main entry point
│   └── bin/             # Platform-specific Go binaries
│       ├── linux_x86_64
│       ├── linux_aarch64
│       ├── macosx_x86_64
│       ├── macosx_arm64
│       ├── win_amd64.exe
│       └── win_arm64.exe
└── README.md            # This document
```

## Development

### Running Tests

```bash
# Run all tests
pytest tests/py_tests/

# Run specific test file
pytest tests/py_tests/test_wrapper.py

# Run with coverage report
pytest tests/py_tests/ --cov=py/gptcomet --cov-report=html

# Verbose output
pytest tests/py_tests/ -v --tb=short
```

### Code Quality

```bash
# Linting
ruff check py/

# Formatting
ruff format py/

# Check format without modifying
ruff format --check py/
```

## Test Files

- `tests/py_tests/test_wrapper.py` - Main Python wrapper tests
- `tests/py_tests/conftest.py` - pytest fixtures and shared test utilities

## Architecture

The Python wrapper serves as a thin layer that:
1. Detects the current platform (Linux, macOS, Windows)
2. Selects the appropriate Go binary
3. Spawns a subprocess to execute the binary
4. Passes through command-line arguments and environment variables
5. Returns the exit code from the Go binary

This design allows:
- Easy installation via `pip install gptcomet`
- Platform-specific binary optimization
- Minimal overhead
- Consistent behavior across platforms
