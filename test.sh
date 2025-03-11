#!/bin/bash
# Repository Heatmap Test Script
# This script is used to run tests

set -e # Stop on errors

# Color definitions
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=========================================================${NC}"
echo -e "${BLUE}        Repository Heatmap Test Script                   ${NC}"
echo -e "${BLUE}=========================================================${NC}"
echo ""

# List of directories to test
TEST_DIRS=(
  "./internal/utils"
  "./internal/heatmap"
)

# Run individual test package
run_test() {
  local package=$1
  echo -e "${BLUE}Running tests: ${package}${NC}"
  
  if go test -v "$package"; then
    echo -e "${GREEN}✓ Test succeeded: ${package}${NC}"
    return 0
  else
    echo -e "${RED}✗ Test failed: ${package}${NC}"
    return 1
  fi
}

# Generate coverage report
generate_coverage() {
  echo -e "${BLUE}Generating coverage report...${NC}"
  
  go test ./internal/... -coverprofile=coverage.out
  go tool cover -html=coverage.out -o coverage.html
  
  echo -e "${GREEN}Coverage report generated at coverage.html${NC}"
}

# Main process
main() {
  local failed=0
  
  # Run tests for each directory
  for dir in "${TEST_DIRS[@]}"; do
    if ! run_test "$dir"; then
      failed=1
    fi
    echo ""
  done
  
  # Generate coverage report
  generate_coverage
  
  # Display results
  echo ""
  echo -e "${BLUE}=========================================================${NC}"
  if [ $failed -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
  else
    echo -e "${RED}✗ Some tests failed${NC}"
  fi
  echo -e "${BLUE}=========================================================${NC}"
  
  return $failed
}

# Run the script
main 