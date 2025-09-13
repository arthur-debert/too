#!/usr/bin/env bash

set -euo pipefail

# Get the project root directory (where this script is located)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
GRAY='\033[38;5;245m' # Medium gray
NC='\033[0m'          # No Color

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
  -i, --interactive     Stay in interactive shell after running script
  -h, --help            Show this help message and exit

Description:
  This script creates a temporary directory with an isolated .todos file for
  testing too commands. It always initializes a fresh .todos file. If a script
  is provided, it executes the script in the test environment and exits
  (unless --interactive is used). Otherwise, it drops you into a new shell
  session where you can run too commands without affecting your real todo data.
  The temporary environment is automatically cleaned up when you exit.

  Note: For the --format option to work with test scripts, the script must use
  --format \${TOO_FORMAT} in all too commands. The TOO_FORMAT environment
  variable is automatically set based on the --format option.

  Note 2: There is a export_history functin that outptus the session history sans line numbers, useful if you want to capture the commands you ran in the session for later use in a script.
Examples:
  $(basename "$0") # Start with a fresh .todos file in interactive mode
  $(basename "$0") test-script.sh # Execute a test script in the environment
  $(basename "$0") --interactive test-script.sh # Execute a script then stay in interactive mode


EOF
}

# Parse arguments
SCRIPT_FILE=""
TOO_FORMAT="term"
INTERACTIVE=false
while [[ $# -gt 0 ]]; do
    case $1 in
    -h | --help)
        usage
        exit 0
        ;;
    -i | --interactive)
        INTERACTIVE=true
        shift
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
        if [[ -z "${SCRIPT_FILE}" ]]; then
            SCRIPT_FILE="${1}"
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
if [[ -n "${SCRIPT_FILE}" ]]; then
    if [[ ! -f "${SCRIPT_FILE}" ]]; then
        echo -e "${RED}Error: Script file '${SCRIPT_FILE}' does not exist${NC}" >&2
        exit 1
    fi
    # Convert to absolute path
    script_dir="$(dirname "${SCRIPT_FILE}")"
    script_name="$(basename "${SCRIPT_FILE}")"
    if ! cd "${script_dir}"; then
        echo -e "${RED}Error: Cannot access directory containing script file${NC}" >&2
        exit 1
    fi
    script_abs_dir="$(pwd)"
    cd - >/dev/null || exit 1
    SCRIPT_FILE="${script_abs_dir}/${script_name}"
    # Make sure it's executable
    if [[ ! -x "${SCRIPT_FILE}" ]]; then
        echo -e "${GRAY}Making script file executable...${NC}"
        chmod +x "${SCRIPT_FILE}"
    fi
fi

# Ensure tmp directory exists
TMP_BASE="${PROJECT_ROOT}/tmp"
mkdir -p "${TMP_BASE}"

# Create a unique temporary directory
TEMP_DIR=$(mktemp -d "${TMP_BASE}/too-test-XXXXXX")

# Always build a fresh binary to ensure we're testing the latest code
TOO_BIN="${PROJECT_ROOT}/bin/too"
echo -e "${GRAY}Building fresh too binary...${NC}"
if ! (cd "${PROJECT_ROOT}" && ./scripts/build --skip-tests >/dev/null 2>&1); then
    echo -e "${RED}Error: Binary build failed${NC}" >&2
    exit 1
fi

# Verify the binary exists and is executable
if [[ ! -x "${TOO_BIN}" ]]; then
    echo -e "${RED}Error: Binary build failed or binary not found at ${TOO_BIN}${NC}" >&2
    exit 1
fi

# Function to cleanup on exit
cleanup() {
    echo -e "\n${YELLOW}Cleaning up test environment...${NC}"
    if [[ -d "${TEMP_DIR}" ]]; then
        rm -rf "${TEMP_DIR}"
        echo -e "${GREEN}Test environment cleaned up${NC}"
    fi
}

# Set up trap to cleanup on exit
trap cleanup EXIT INT TERM

# Change to temp directory
cd "${TEMP_DIR}"

# Always initialize a fresh .todos file
echo -e "${GRAY}Initializing fresh .todos file${NC}"
"${TOO_BIN}" init >/dev/null 2>&1

# Get version info
VERSION_INFO=$("${TOO_BIN}" --version 2>&1 || echo "Version unknown")

# Display environment info
echo -e "${GRAY}Working directory: ${TEMP_DIR}${NC}"
echo -e "${GRAY}Too version: ${VERSION_INFO}${NC}"
echo -e "${GRAY}Output format: ${TOO_FORMAT}${NC}"

# Add project bin to PATH for easy too access
export PATH="${PROJECT_ROOT}/bin:${PATH}"

# Export the format for use in scripts
export TOO_FORMAT="${TOO_FORMAT}"

# Set up history file for the session
export HISTFILE="${TEMP_DIR}/.zsh_history"
export HISTSIZE=10000
export SAVEHIST=10000

# Create a function to export history without line numbers
function export_history() {
    fc -ln 1 2>/dev/null || echo "# No history available"
}

# Function to launch interactive shell or execute script
launch_shell() {
    local script_to_run="${1:-}"
    local start_interactive_shell=false

    if [[ -n "${script_to_run}" ]]; then
        # Execute the script
        echo -e "${GRAY}Executing script: ${script_to_run}${NC}\n"

        # Use zsh to execute the script in the current environment
        if ! command -v zsh >/dev/null 2>&1; then
            echo -e "${RED}Error: zsh not found${NC}" >&2
            exit 1
        fi

        # Execute the script with the current environment
        # Extract commands from the script (excluding comments and empty lines)
        if ! grep -v '^#' "${script_to_run}" | grep -v '^$' >"${TEMP_DIR}/.command_history"; then
            true # Ignore grep failures
        fi

        # Create a wrapper script that includes our functions
        cat >"${TEMP_DIR}/.wrapper.zsh" <<EOF
#!/usr/bin/env zsh

# Define export_history function that reads captured commands
function export_history() {
    if [[ -f "${TEMP_DIR}/.command_history" ]]; then
        grep -v 'export_history' "${TEMP_DIR}/.command_history" 2>/dev/null || true
    else
        echo "# No history available"
    fi
}

# Read and execute the script line by line
while IFS= read -r line; do
    # Skip comments and empty lines
    if [[ "\$line" =~ ^[[:space:]]*# ]] || [[ -z "\$line" ]]; then
        continue
    fi
    
    # Echo the command in gray
    echo 
    echo -e "${GRAY}$ \$line${NC}"
    
    # Execute the command
    eval "\$line"
done < "\$1"
EOF
        chmod +x "${TEMP_DIR}/.wrapper.zsh"

        # Run zsh with no rc files to avoid user config interference
        zsh --no-globalrcs --no-rcs "${TEMP_DIR}/.wrapper.zsh" "${script_to_run}"

        # Check if we should start interactive shell after script
        if [[ "${INTERACTIVE}" == "true" ]]; then
            start_interactive_shell=true
        fi
    else
        start_interactive_shell=true
    fi

    if [[ "${start_interactive_shell}" == "true" ]]; then
        # Start interactive shell
        echo -e "\n${YELLOW}Type 'exit' or press Ctrl+D to leave and cleanup${NC}\n"

        # Create a custom prompt to indicate test environment
        export PS1="[too-test] $ "

        # Create zshrc with vi mode
        cat >"${TEMP_DIR}/.zshrc" <<'EOF'
# Set vi mode
bindkey -v

# Set custom prompt
PS1="[too-test] $ "
EOF

        # Start a new shell with custom config
        ZDOTDIR="${TEMP_DIR}" zsh
    fi
}

# Launch shell or execute script
launch_shell "${SCRIPT_FILE}"
