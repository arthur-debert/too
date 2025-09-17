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

# Default output format
OUTPUT_FORMAT="nice"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --output)
            OUTPUT_FORMAT="$2"
            shift 2
            ;;
        --help|-h)
            echo "Usage: $0 [--output FORMAT]"
            echo ""
            echo "Formats:"
            echo "  nice    Human-friendly output with colors and timing (default)"
            echo "  junit   JUnit XML format for CI systems"
            echo "  tap     TAP (Test Anything Protocol) format"
            echo ""
            echo "Examples:"
            echo "  $0                    # Run with nice output"
            echo "  $0 --output junit     # Generate JUnit XML"
            echo "  $0 --output tap       # Generate TAP output"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Validate output format
case "$OUTPUT_FORMAT" in
    nice|junit|tap)
        ;;
    *)
        echo "Error: Invalid output format '$OUTPUT_FORMAT'"
        echo "Valid formats: nice, junit, tap"
        exit 1
        ;;
esac

echo -e "${BLUE}🧪 Running too e2e tests...${NC}"

# Run tests based on selected format
case "$OUTPUT_FORMAT" in
    nice)
        echo -e "${YELLOW}📊 Running tests with human-friendly output:${NC}"
        if bats --pretty --timing "$SCRIPT_DIR"/suite/*.bats; then
            echo -e "${GREEN}✅ All tests passed!${NC}"
            EXIT_CODE=0
        else
            echo -e "${YELLOW}⚠️  Some tests failed${NC}"
            EXIT_CODE=1
        fi
        ;;
    junit)
        # Create results directory only for junit output
        if [ -n "${E2E_RESULTS_DIR:-}" ]; then
            RESULTS_DIR="$E2E_RESULTS_DIR"
        else
            RESULTS_DIR=$(mktemp -d -t too-e2e-results.XXXXXX)
        fi
        mkdir -p "$RESULTS_DIR"
        
        echo -e "${YELLOW}📝 Generating JUnit XML report...${NC}"
        if bats --formatter junit "$SCRIPT_DIR"/suite/*.bats > "$RESULTS_DIR/junit.xml"; then
            echo -e "${GREEN}✅ All tests passed!${NC}"
            echo -e "${BLUE}📁 Results saved to: $RESULTS_DIR/junit.xml${NC}"
            EXIT_CODE=0
        else
            echo -e "${YELLOW}⚠️  Some tests failed${NC}"
            echo -e "${BLUE}📁 Results saved to: $RESULTS_DIR/junit.xml${NC}"
            EXIT_CODE=1
        fi
        ;;
    tap)
        echo -e "${YELLOW}📋 Generating TAP report...${NC}"
        if bats --tap "$SCRIPT_DIR"/suite/*.bats; then
            EXIT_CODE=0
        else
            EXIT_CODE=1
        fi
        ;;
esac

exit $EXIT_CODE