#!/bin/bash

# Run integration tests only

set -e

echo "Running integration tests..."
echo "==========================="

# Set test environment variables
export WASTEBIN_LOCAL_DB=true
export WASTEBIN_LOG_LEVEL=ERROR

# Run tests with integration tag
go test -v -tags=integration ./... 2>&1 | grep -v "no test files" || true

echo ""
echo "Integration tests completed!"