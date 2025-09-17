#!/bin/bash

# Run all e2e test suites
# This script runs all Bats test files in the suite directory

set -euo pipefail

# Get the directory containing this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
E2E_DIR="$(dirname "$SCRIPT_DIR")"

echo "🧪 Running too e2e test suite..."
echo "================================"

# Run the e2e test runner which handles all .bats files
"$E2E_DIR/run-tests.sh"

echo "================================"
echo "✅ E2E test suite complete!"