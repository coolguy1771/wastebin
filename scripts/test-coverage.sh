#!/bin/bash

# Test coverage script for Wastebin
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
COVERAGE_THRESHOLD=75
COVERAGE_FILE="coverage.out"
COVERAGE_HTML="coverage.html"

echo -e "${GREEN}🧪 Running Wastebin Test Suite${NC}"
echo "=================================="

# Set environment variables for testing
export WASTEBIN_LOCAL_DB=true
export WASTEBIN_LOG_LEVEL=ERROR

# Clean up previous coverage files
rm -f ${COVERAGE_FILE} ${COVERAGE_HTML}

# Run tests with coverage
echo -e "${YELLOW}Running unit tests with coverage...${NC}"
go test -v -race -covermode=atomic -coverprofile=${COVERAGE_FILE} ./...

if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Unit tests failed${NC}"
    exit 1
fi

# Run integration tests if not in short mode
if [ "${1}" != "--short" ]; then
    echo -e "${YELLOW}Running integration tests...${NC}"
    go test -v -tags=integration ./tests/...
    
    if [ $? -ne 0 ]; then
        echo -e "${RED}❌ Integration tests failed${NC}"
        exit 1
    fi
fi

# Generate coverage report
if [ -f ${COVERAGE_FILE} ]; then
    echo -e "${YELLOW}Generating coverage report...${NC}"
    
    # Calculate total coverage
    COVERAGE=$(go tool cover -func=${COVERAGE_FILE} | grep total | grep -oE '[0-9]+\.[0-9]+')
    
    echo "Coverage: ${COVERAGE}%"
    
    # Check if coverage meets threshold
    if (( $(echo "${COVERAGE} >= ${COVERAGE_THRESHOLD}" | bc -l) )); then
        echo -e "${GREEN}✅ Coverage threshold met (${COVERAGE}% >= ${COVERAGE_THRESHOLD}%)${NC}"
    else
        echo -e "${RED}❌ Coverage below threshold (${COVERAGE}% < ${COVERAGE_THRESHOLD}%)${NC}"
        exit 1
    fi
    
    # Generate HTML report
    go tool cover -html=${COVERAGE_FILE} -o ${COVERAGE_HTML}
    echo "HTML coverage report generated: ${COVERAGE_HTML}"
    
    # Show per-package coverage
    echo -e "${YELLOW}Per-package coverage:${NC}"
    go tool cover -func=${COVERAGE_FILE} | grep -v total | sort -k3 -nr
    
else
    echo -e "${RED}❌ Coverage file not found${NC}"
    exit 1
fi

# Run benchmarks if requested
if [ "${1}" == "--bench" ] || [ "${2}" == "--bench" ]; then
    echo -e "${YELLOW}Running benchmarks...${NC}"
    go test -bench=. -benchmem -run=^$ ./... | tee benchmark.txt
fi

# Run security tests
echo -e "${YELLOW}Running security tests...${NC}"
go test -v ./tests/ -run TestSecurity

# Final summary
echo ""
echo -e "${GREEN}🎉 All tests completed successfully!${NC}"
echo "Coverage: ${COVERAGE}%"
echo "HTML Report: ${COVERAGE_HTML}"

# Optional: Open coverage report in browser (uncomment if desired)
# if command -v open >/dev/null 2>&1; then
#     open ${COVERAGE_HTML}
# elif command -v xdg-open >/dev/null 2>&1; then
#     xdg-open ${COVERAGE_HTML}
# fi