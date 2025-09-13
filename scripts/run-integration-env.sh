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
    cat <<EOF
Usage: $(basename "$0") [OPTIONS] [SCRIPT_FILE]

Create an isolated testing environment for too commands with a temporary data store.

Arguments:
  SCRIPT_FILE   Optional shell script to execute in the test environment.
                If not provided, starts an interactive shell session.

Options:
  -f, --format FORMAT   Set output format for too commands (term|json|markdown)
                        Default: term
  -h, --help            Show this help message and exit

Description:
  This script creates a temporary directory with an isolated .todos file for
  testing too commands. It always initializes a fresh .todos file. If a script
  is provided, it executes the script in the test environment and exits.
  Otherwise, it drops you into a new shell session where you can run too
  commands without affecting your real todo data. The temporary environment
  is automatically cleaned up when you exit.

  Note: For the --format option to work with test scripts, the script must use
  --format \${TOO_FORMAT} in all too commands. The TOO_FORMAT environment
  variable is automatically set based on the --format option.

Examples:
  # Start with a fresh .todos file in interactive mode
  $(basename "$0")

  # Execute a test script in the environment
  $(basename "$0") test-script.sh

  # Execute a script with JSON format
  $(basename "$0") --format json test-script.sh

  # Execute a script with relative path
  $(basename "$0") -f markdown ../tests/integration-test.sh

Example test script that uses TOO_FORMAT:
  #!/bin/bash
  echo "Testing with format: \${TOO_FORMAT}"
  too add "My task" --format "\${TOO_FORMAT}"
  too list --format "\${TOO_FORMAT}"

EOF
}

# Parse arguments
SCRIPT_FILE=""
TOO_FORMAT="term"
while [[ $# -gt 0 ]]; do
    case $1 in
    -h | --help)
        usage
        exit 0
        ;;
    -f | --format)
        if [[ -n "${2:-}" ]]; then
            case "${2}" in
            term | json | markdown)
                TOO_FORMAT="${2}"
                shift 2
                ;;
            *)
                echo -e "${RED}Error: Invalid format '${2}'. Valid formats are: term, json, markdown${NC}" >&2
                exit 1
                ;;
            esac
        else
            echo -e "${RED}Error: --format requires a value${NC}" >&2
            exit 1
        fi
        ;;
    --format=*)
        # Handle --format=VALUE format
        value="${1#*=}"
        case "${value}" in
        term | json | markdown)
            TOO_FORMAT="${value}"
            shift
            ;;
        *)
            echo -e "${RED}Error: Invalid format '${value}'. Valid formats are: term, json, markdown${NC}" >&2
            exit 1
            ;;
        esac
        ;;
    *)
        if [[ -z "$SCRIPT_FILE" ]]; then
            SCRIPT_FILE="$1"
        else
            echo -e "${RED}Error: Too many arguments${NC}" >&2
            usage
            exit 1
        fi
        shift
        ;;
    esac
done

# Validate script file if provided
if [[ -n "$SCRIPT_FILE" ]]; then
    if [[ ! -f "$SCRIPT_FILE" ]]; then
        echo -e "${RED}Error: Script file '$SCRIPT_FILE' does not exist${NC}" >&2
        exit 1
    fi
    # Convert to absolute path
    SCRIPT_FILE="$(cd "$(dirname "$SCRIPT_FILE")" && pwd)/$(basename "$SCRIPT_FILE")"
    # Make sure it's executable
    if [[ ! -x "$SCRIPT_FILE" ]]; then
        echo -e "${YELLOW}Making script file executable...${NC}"
        chmod +x "$SCRIPT_FILE"
    fi
fi

# Ensure tmp directory exists
TMP_BASE="$PROJECT_ROOT/tmp"
mkdir -p "$TMP_BASE"

# Create a unique temporary directory
TEMP_DIR=$(mktemp -d "$TMP_BASE/too-test-XXXXXX")

# Build too if needed
TOO_BIN="$PROJECT_ROOT/bin/too"
if [[ ! -x "$TOO_BIN" ]]; then
    echo -e "${YELLOW}Building too...${NC}"
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

# Always initialize a fresh .todos file
echo -e "${BLUE}Initializing fresh .todos file${NC}"
"$TOO_BIN" init

# Display environment info
echo -e "\n${GREEN}=== Test Environment Ready ===${NC}"
echo -e "Working directory: ${BLUE}$TEMP_DIR${NC}"
echo -e "Using too binary: ${BLUE}$TOO_BIN${NC}"
echo -e "Data file: ${BLUE}$TEMP_DIR/.todos${NC}"
echo -e "Output format: ${BLUE}$TOO_FORMAT${NC}"

# Add project bin to PATH for easy too access
export PATH="$PROJECT_ROOT/bin:$PATH"

# Export the format for use in scripts
export TOO_FORMAT="$TOO_FORMAT"

# Function to launch interactive shell or execute script
launch_shell() {
    local script_to_run="${1:-}"

    if [[ -n "${script_to_run}" ]]; then
        # Execute the script and exit
        echo -e "${YELLOW}Executing script: ${script_to_run}${NC}\n"

        # Use bash to execute the script in the current environment
        if ! command -v bash >/dev/null 2>&1; then
            echo -e "${RED}Error: bash not found${NC}" >&2
            exit 1
        fi

        # Execute the script with the current environment
        bash "${script_to_run}"
    else
        # Start interactive shell
        echo -e "\n${YELLOW}Type 'exit' or press Ctrl+D to leave and cleanup${NC}\n"

        # Create a custom prompt to indicate test environment
        export PS1="[too-test] \w $ "

        # Start a new shell (without exec to ensure cleanup trap runs)
        bash --norc
    fi
}

# Launch shell or execute script
launch_shell "${SCRIPT_FILE}"
