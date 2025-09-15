#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get the baseline directory
BASELINE_DIR="live-tests/baseline-latest"
if [ ! -d "$BASELINE_DIR" ]; then
    echo -e "${RED}No baseline found. Run capture-baseline.sh first.${NC}"
    exit 1
fi

# Create new output directory
NEW_DIR="live-tests/output-$(date +%Y%m%d-%H%M%S)"
mkdir -p "$NEW_DIR"

echo "Comparing against baseline: $(readlink $BASELINE_DIR)"
echo "New output in: $NEW_DIR"

# First, build the current version
echo -e "\n${YELLOW}Building too...${NC}"
./scripts/build || exit 1

# Track differences
DIFFERENCES=0

# Run each test script and compare
for script in live-tests/*.sh; do
    if [ -f "$script" ] && [ -x "$script" ]; then
        script_name=$(basename "$script" .sh)
        baseline_file="$BASELINE_DIR/${script_name}.out"
        new_file="$NEW_DIR/${script_name}.out"
        
        if [ ! -f "$baseline_file" ]; then
            echo -e "${YELLOW}Skipping $script_name (no baseline)${NC}"
            continue
        fi
        
        echo -e "\n${YELLOW}Testing $script_name...${NC}"
        
        # Run the script and capture output
        ./live-tests/run "$script" > "$new_file" 2>&1
        
        # Compare outputs, ignoring timestamps and other variable content
        if diff -q \
            <(sed -E 's/[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}[^ ]*/TIMESTAMP/g' "$baseline_file" | \
              sed -E 's/built: [0-9]{4}-[0-9]{2}-[0-9]{2}_[0-9]{2}:[0-9]{2}:[0-9]{2}/built: TIMESTAMP/g' | \
              sed -E 's/baseline-[0-9]{8}-[0-9]{6}/baseline-TIMESTAMP/g' | \
              sed -E 's/output-[0-9]{8}-[0-9]{6}/output-TIMESTAMP/g' | \
              sed -E 's/too-test-[a-zA-Z0-9]{6}/too-test-RANDOM/g' | \
              sed -E 's/Too version: .* \(commit: [a-f0-9]+, built: TIMESTAMP\)/Too version: VERSION/g') \
            <(sed -E 's/[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}[^ ]*/TIMESTAMP/g' "$new_file" | \
              sed -E 's/built: [0-9]{4}-[0-9]{2}-[0-9]{2}_[0-9]{2}:[0-9]{2}:[0-9]{2}/built: TIMESTAMP/g' | \
              sed -E 's/baseline-[0-9]{8}-[0-9]{6}/baseline-TIMESTAMP/g' | \
              sed -E 's/output-[0-9]{8}-[0-9]{6}/output-TIMESTAMP/g' | \
              sed -E 's/too-test-[a-zA-Z0-9]{6}/too-test-RANDOM/g' | \
              sed -E 's/Too version: .* \(commit: [a-f0-9]+, built: TIMESTAMP\)/Too version: VERSION/g') > /dev/null; then
            echo -e "  ${GREEN}✓ Output matches baseline${NC}"
        else
            echo -e "  ${RED}✗ Output differs from baseline${NC}"
            DIFFERENCES=$((DIFFERENCES + 1))
            
            # Save the diff for inspection
            diff -u \
                <(sed -E 's/[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}[^ ]*/TIMESTAMP/g' "$baseline_file") \
                <(sed -E 's/[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}[^ ]*/TIMESTAMP/g' "$new_file") \
                > "$NEW_DIR/${script_name}.diff"
            
            echo "    Diff saved to: $NEW_DIR/${script_name}.diff"
            echo "    To see the diff: diff -u $baseline_file $new_file"
        fi
    fi
done

echo -e "\n========================================="
if [ $DIFFERENCES -eq 0 ]; then
    echo -e "${GREEN}All tests match baseline! ✨${NC}"
else
    echo -e "${RED}Found $DIFFERENCES differences from baseline${NC}"
    echo "Review diffs in: $NEW_DIR/*.diff"
fi