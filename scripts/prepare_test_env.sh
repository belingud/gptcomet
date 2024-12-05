#!/bin/bash

# Set version
VERSION="0.0.24"
BINARY_NAME="gptcomet"

# Create test directories
mkdir -p test_dist/py/gptcomet

# Build Go binary for current platform
echo "Building Go binary..."
go build -o ${BINARY_NAME} .

# Copy binary to Python package directory with correct naming
GOOS=$(uname -s | tr '[:upper:]' '[:lower:]')
if [[ "$(uname -m)" == "x86_64" ]]; then
    GOARCH="amd64"
elif [[ "$(uname -m)" == "arm64" ]]; then
    GOARCH="arm64"
else
    GOARCH="$(uname -m)"
fi

# Copy binary with platform-specific name
if [[ "$GOOS" == "darwin" ]]; then
    cp ${BINARY_NAME} test_dist/py/gptcomet/${BINARY_NAME}_darwin_${GOARCH}
elif [[ "$GOOS" == "linux" ]]; then
    cp ${BINARY_NAME} test_dist/py/gptcomet/${BINARY_NAME}_linux_${GOARCH}
fi

# Copy Python package files
cp -r py/* test_dist/py/

# Clean up
rm ${BINARY_NAME}

echo "Test environment prepared in test_dist/py/"
echo "To test the package, run:"
echo "cd test_dist/py && pdm install && pdm build"
