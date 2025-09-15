#!/bin/bash

# Create baseline directory with timestamp
BASELINE_DIR="live-tests/baseline-$(date +%Y%m%d-%H%M%S)"
mkdir -p "$BASELINE_DIR"

echo "Capturing baseline output to: $BASELINE_DIR"

# First, build the current version
echo "Building too..."
./scripts/build || exit 1

# Run each test script and capture output
for script in live-tests/*.sh; do
    if [ -f "$script" ] && [ -x "$script" ]; then
        script_name=$(basename "$script" .sh)
        echo "Running $script_name..."
        
        # Run the script and capture both stdout and stderr
        ./live-tests/run "$script" > "$BASELINE_DIR/${script_name}.out" 2>&1
        
        # Also capture just the commands and their output (no shell prompts)
        ./live-tests/run "$script" 2>&1 | grep -E "^\+" | sed 's/^+ /$ /' > "$BASELINE_DIR/${script_name}.commands"
    fi
done

# Create a symlink to the latest baseline
ln -sfn "$(basename "$BASELINE_DIR")" live-tests/baseline-latest

echo "Baseline captured in: $BASELINE_DIR"
echo "Symlinked as: live-tests/baseline-latest"