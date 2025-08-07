#!/bin/bash

# Run unit tests only (without integration tests)

set -e

echo "Running unit tests..."
echo "===================="

# Set test environment variables
export WASTEBIN_LOCAL_DB=true
export WASTEBIN_LOG_LEVEL=ERROR

# Run tests without integration tag
go test -v -race -short ./... 2>&1 | grep -v "no test files" || true

echo ""
echo "Unit tests completed!"