#!/usr/bin/env bash

set -euo pipefail

# Get the project root directory (where this script is located)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Usage function
usage() {
    cat << EOF
Usage: $(basename "$0") [OPTIONS] [TODOS_FILE]

Create an isolated testing environment for tdh commands with a temporary data store.

Arguments:
  TODOS_FILE    Optional .todos file to populate the test environment with.
                If not provided, a fresh .todos file will be initialized.

Options:
  -h, --help    Show this help message and exit

Description:
  This script creates a temporary directory with an isolated .todos file for
  testing tdh commands. It drops you into a new shell session where you can
  run tdh commands without affecting your real todo data. The temporary
  environment is automatically cleaned up when you exit the shell.

Examples:
  # Start with a fresh .todos file
  $(basename "$0")

  # Start with a copy of an existing .todos file
  $(basename "$0") ~/.todos.json

  # Start with a test data file
  $(basename "$0") test/fixtures/sample.todos.json

EOF
}

# Parse arguments
TODOS_FILE=""
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            usage
            exit 0
            ;;
        *)
            if [[ -z "$TODOS_FILE" ]]; then
                TODOS_FILE="$1"
            else
                echo -e "${RED}Error: Too many arguments${NC}" >&2
                usage
                exit 1
            fi
            shift
            ;;
    esac
done

# Validate todos file if provided
if [[ -n "$TODOS_FILE" ]]; then
    if [[ ! -f "$TODOS_FILE" ]]; then
        echo -e "${RED}Error: File '$TODOS_FILE' does not exist${NC}" >&2
        exit 1
    fi
    # Convert to absolute path
    TODOS_FILE="$(cd "$(dirname "$TODOS_FILE")" && pwd)/$(basename "$TODOS_FILE")"
fi

# Ensure tmp directory exists
TMP_BASE="$PROJECT_ROOT/tmp"
mkdir -p "$TMP_BASE"

# Create a unique temporary directory
TEMP_DIR=$(mktemp -d "$TMP_BASE/tdh-test-XXXXXX")

# Build tdh if needed
TDH_BIN="$PROJECT_ROOT/bin/tdh"
if [[ ! -x "$TDH_BIN" ]]; then
    echo -e "${YELLOW}Building tdh...${NC}"
    (cd "$PROJECT_ROOT" && ./scripts/build)
fi

# Function to cleanup on exit
cleanup() {
    echo -e "\n${YELLOW}Cleaning up test environment...${NC}"
    if [[ -d "$TEMP_DIR" ]]; then
        rm -rf "$TEMP_DIR"
        echo -e "${GREEN}Test environment cleaned up${NC}"
    fi
}

# Set up trap to cleanup on exit
trap cleanup EXIT INT TERM

# Change to temp directory
cd "$TEMP_DIR"

# Initialize or copy todos file
if [[ -n "$TODOS_FILE" ]]; then
    echo -e "${BLUE}Copying todos file from: $TODOS_FILE${NC}"
    cp "$TODOS_FILE" .todos
    echo -e "${GREEN}Test environment initialized with existing data${NC}"
else
    echo -e "${BLUE}Initializing fresh .todos file${NC}"
    "$TDH_BIN" init
fi

# Display environment info
echo -e "\n${GREEN}=== Test Environment Ready ===${NC}"
echo -e "Working directory: ${BLUE}$TEMP_DIR${NC}"
echo -e "Using tdh binary: ${BLUE}$TDH_BIN${NC}"
echo -e "Data file: ${BLUE}$TEMP_DIR/.todos${NC}"
echo -e "\n${YELLOW}Type 'exit' or press Ctrl+D to leave and cleanup${NC}\n"

# Create a custom prompt to indicate test environment
export PS1="[tdh-test] \w $ "

# Add project bin to PATH for easy tdh access
export PATH="$PROJECT_ROOT/bin:$PATH"

# Start a new shell
exec bash --norc