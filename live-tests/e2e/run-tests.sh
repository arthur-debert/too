#!/bin/bash

set -euo pipefail

# Set TERM for CI environments
export TERM=${TERM:-xterm}

# Get the directory containing this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}🧪 Running too e2e tests...${NC}"

# Create results directory
RESULTS_DIR="$SCRIPT_DIR/results"
mkdir -p "$RESULTS_DIR"

# Clean up old results
rm -f "$RESULTS_DIR"/*.xml "$RESULTS_DIR"/*.tap

# Run tests with pretty output (for humans)
echo -e "${YELLOW}📊 Running tests with human-friendly output:${NC}"
if bats --pretty --timing "$SCRIPT_DIR"/suite/*.bats; then
    echo -e "${GREEN}✅ All tests passed!${NC}"
    EXIT_CODE=0
else
    echo -e "${YELLOW}⚠️  Some tests failed${NC}"
    EXIT_CODE=1
fi

# Run tests with JUnit XML output (for CI)
echo -e "${YELLOW}📝 Generating JUnit XML report...${NC}"
bats --formatter junit "$SCRIPT_DIR"/suite/*.bats > "$RESULTS_DIR/junit.xml" || true

# Run tests with TAP output (for compatibility)
echo -e "${YELLOW}📋 Generating TAP report...${NC}"
bats --tap --output "$RESULTS_DIR" "$SCRIPT_DIR"/suite/*.bats > "$RESULTS_DIR/tests.tap" || true

echo -e "${BLUE}📁 Results saved to: $RESULTS_DIR${NC}"
ls -la "$RESULTS_DIR"

exit $EXIT_CODE