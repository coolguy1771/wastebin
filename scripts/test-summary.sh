#!/bin/bash

# Show test summary for the project

set -e

echo "Wastebin Test Summary"
echo "===================="
echo ""

# Count test files
unit_tests=$(find . -name "*_test.go" -not -name "*_integration_test.go" -not -path "./vendor/*" -not -path "./web/*" | wc -l | tr -d ' ')
integration_tests=$(find . -name "*_integration_test.go" -not -path "./vendor/*" -not -path "./web/*" | wc -l | tr -d ' ')

echo "Test Files:"
echo "  Unit test files:        $unit_tests"
echo "  Integration test files: $integration_tests"
echo ""

# Count test functions
echo "Test Functions:"
unit_funcs=$(grep -h "^func Test" "$(find . -name "*_test.go" -not -name "*_integration_test.go" -not -path "./vendor/*" -not -path "./web/*" 2>/dev/null)" 2>/dev/null | wc -l | tr -d ' ')
integration_funcs=$(grep -h "^func Test" "$(find . -name "*_integration_test.go" -not -path "./vendor/*" -not -path "./web/*" 2>/dev/null)" 2>/dev/null | wc -l | tr -d ' ')

echo "  Unit test functions:        $unit_funcs"
echo "  Integration test functions: $integration_funcs"
echo ""

# Show package breakdown
echo "Package Breakdown:"
echo ""
echo "Unit Tests by Package:"
for pkg in $(go list ./... | grep -v /vendor/); do
    count=$(go test -list=. $pkg 2>/dev/null | grep -c "^Test" || true)
    if [ -n "$count" ] && [ "$count" -gt 0 ]; then
        echo "  $pkg: $count tests"
    fi
done

echo ""
echo "Integration Tests by Package:"
while IFS= read -r -d '' file; do
    pkg=$(dirname "$file")
    count=$(grep -c "^func Test" "$file" || echo 0)
    if [ "$count" -gt 0 ]; then
        echo "  $pkg: $count tests"
    fi
done < <(find . -name "*_integration_test.go" -not -path "./vendor/*" -not -path "./web/*" -print0)

echo ""
echo "To run specific test types:"
echo "  Unit tests:        make test"
echo "  Integration tests: make test-integration"
echo "  All tests:         make test-all"