#!/usr/bin/env bats

# Test suite for contextual view feature
# Tests: --contextual flag functionality, ellipsis display, nested item contextual display

setup() {
    # Get the directory containing this test file
    TEST_DIR="$( cd "$( dirname "$BATS_TEST_FILENAME" )" && pwd )"
    E2E_DIR="$(dirname "$TEST_DIR")"
    LIVE_TESTS_DIR="$(dirname "$E2E_DIR")"
    PROJECT_ROOT="$(dirname "$LIVE_TESTS_DIR")"
    
    # Load utility functions
    source "$E2E_DIR/utils/parse-json.sh"
    
    # Export paths for use in tests
    export TEST_DIR
    export E2E_DIR
    export LIVE_TESTS_DIR
    export PROJECT_ROOT
    
    # Create a temporary directory for this test
    export TEST_TEMP_DIR="$(mktemp -d)"
    export TODO_DB_PATH="$TEST_TEMP_DIR/.todos.json"
    
    # Change to temp directory
    cd "$TEST_TEMP_DIR"
    
    # Build the too binary if needed
    if [[ ! -f "$PROJECT_ROOT/bin/too" ]]; then
        (cd "$PROJECT_ROOT" && go build -o bin/too ./cmd/too)
    fi
    
    # Create alias for too with proper data path
    too() {
        "$PROJECT_ROOT/bin/too" --data-path="$TODO_DB_PATH" "$@"
    }
    export -f too
}

teardown() {
    # Clean up temporary directory
    if [[ -n "$TEST_TEMP_DIR" && -d "$TEST_TEMP_DIR" ]]; then
        rm -rf "$TEST_TEMP_DIR"
    fi
}

# Helper function to check output contains text
assert_output_contains() {
    local expected="$1"
    if [[ "$output" != *"$expected"* ]]; then
        echo "Expected output to contain: $expected"
        echo "Actual output: $output"
        return 1
    fi
}

# Helper function to check output does not contain text
refute_output_contains() {
    local unexpected="$1"
    if [[ "$output" == *"$unexpected"* ]]; then
        echo "Expected output NOT to contain: $unexpected"
        echo "Actual output: $output"
        return 1
    fi
}

@test "contextual view shows reduced list with ellipsis" {
    too init >/dev/null 2>&1
    
    # Create 10 items (suppress output)
    too add "Task 1" >/dev/null 2>&1
    too add "Task 2" >/dev/null 2>&1
    too add "Task 3" >/dev/null 2>&1
    too add "Task 4" >/dev/null 2>&1
    too add "Task 5" >/dev/null 2>&1
    too add "Task 6" >/dev/null 2>&1
    too add "Task 7" >/dev/null 2>&1
    too add "Task 8" >/dev/null 2>&1
    too add "Task 9" >/dev/null 2>&1
    too add "Task 10" >/dev/null 2>&1

    # Edit item 8 with contextual view
    run too --contextual edit 8 "Task 8 (updated)"
    [ "$status" -eq 0 ]
    
    # Should show ellipsis before (items 1-5 are hidden)
    assert_output_contains "…"
    
    # Should show contextual items around Task 8
    assert_output_contains "Task 6"
    assert_output_contains "Task 7"
    assert_output_contains "Task 8 (updated)"
    assert_output_contains "Task 9"
    assert_output_contains "Task 10"
    
    # Should NOT show distant items (use specific patterns)
    refute_output_contains "○ 1. Task 1"
    refute_output_contains "○ 2. Task 2"
    refute_output_contains "○ 3. Task 3"
    refute_output_contains "○ 4. Task 4"
    refute_output_contains "○ 5. Task 5"
}

@test "default view shows all items without ellipsis" {
    too init >/dev/null 2>&1
    
    # Create 7 items (suppress output)
    too add "Item 1" >/dev/null 2>&1
    too add "Item 2" >/dev/null 2>&1
    too add "Item 3" >/dev/null 2>&1
    too add "Item 4" >/dev/null 2>&1
    too add "Item 5" >/dev/null 2>&1
    too add "Item 6" >/dev/null 2>&1
    too add "Item 7" >/dev/null 2>&1

    # Edit item 4 without contextual view (default)
    run too edit 4 "Item 4 (changed)"
    [ "$status" -eq 0 ]
    
    # Should show all items
    assert_output_contains "Item 1"
    assert_output_contains "Item 2"
    assert_output_contains "Item 3"
    assert_output_contains "Item 4 (changed)"
    assert_output_contains "Item 5"
    assert_output_contains "Item 6"
    assert_output_contains "Item 7"
    
    # Should NOT show ellipsis
    refute_output_contains "…"
}

@test "contextual view with nested items shows parent path" {
    too init >/dev/null 2>&1
    
    # Create parent and many children (suppress output)
    too add "Project Alpha" >/dev/null 2>&1
    too add --to 1 "Subtask 1" >/dev/null 2>&1
    too add --to 1 "Subtask 2" >/dev/null 2>&1
    too add --to 1 "Subtask 3" >/dev/null 2>&1
    too add --to 1 "Subtask 4" >/dev/null 2>&1
    too add --to 1 "Subtask 5" >/dev/null 2>&1
    too add --to 1 "Subtask 6" >/dev/null 2>&1
    too add --to 1 "Subtask 7" >/dev/null 2>&1
    too add --to 1 "Subtask 8" >/dev/null 2>&1

    # Add nested item with contextual view
    run too --contextual add --to 1 "Subtask 9"
    [ "$status" -eq 0 ]
    
    # Should show parent item
    assert_output_contains "Project Alpha"
    
    # Should show ellipsis for hidden subtasks
    assert_output_contains "…"
    
    # Should show context around the new subtask
    assert_output_contains "Subtask 7"
    assert_output_contains "Subtask 8"
    assert_output_contains "Subtask 9"
    
    # Should NOT show distant subtasks (use specific patterns)
    refute_output_contains "○ 1.1. Subtask 1"
    refute_output_contains "○ 1.2. Subtask 2"
    refute_output_contains "○ 1.3. Subtask 3"
    refute_output_contains "○ 1.4. Subtask 4"
}

@test "contextual view edge case - item at beginning" {
    too init >/dev/null 2>&1
    
    # Create 8 items (suppress output)
    too add "Beginning 1" >/dev/null 2>&1
    too add "Beginning 2" >/dev/null 2>&1
    too add "Beginning 3" >/dev/null 2>&1
    too add "Beginning 4" >/dev/null 2>&1
    too add "Beginning 5" >/dev/null 2>&1
    too add "Beginning 6" >/dev/null 2>&1
    too add "Beginning 7" >/dev/null 2>&1
    too add "Beginning 8" >/dev/null 2>&1

    # Edit first item with contextual view
    run too --contextual edit 1 "Beginning 1 (first)"
    [ "$status" -eq 0 ]
    
    # Should show the edited first item
    assert_output_contains "Beginning 1 (first)"
    
    # Should show 2 items after it
    assert_output_contains "Beginning 2"
    assert_output_contains "Beginning 3"
    
    # Should show ellipsis after (since items 4-8 are hidden)
    assert_output_contains "…"
    
    # Should NOT show distant items (use specific patterns)
    refute_output_contains "○ 6. Beginning 6"
    refute_output_contains "○ 7. Beginning 7"
    refute_output_contains "○ 8. Beginning 8"
}

@test "contextual view edge case - item at end" {
    too init >/dev/null 2>&1
    
    # Create 8 items (suppress output)
    too add "End 1" >/dev/null 2>&1
    too add "End 2" >/dev/null 2>&1
    too add "End 3" >/dev/null 2>&1
    too add "End 4" >/dev/null 2>&1
    too add "End 5" >/dev/null 2>&1
    too add "End 6" >/dev/null 2>&1
    too add "End 7" >/dev/null 2>&1
    too add "End 8" >/dev/null 2>&1

    # Edit last item with contextual view
    run too --contextual edit 8 "End 8 (last)"
    [ "$status" -eq 0 ]
    
    # Should show the edited last item
    assert_output_contains "End 8 (last)"
    
    # Should show 2 items before it
    assert_output_contains "End 6"
    assert_output_contains "End 7"
    
    # Should show ellipsis before (since items 1-5 are hidden)
    assert_output_contains "…"
    
    # Should NOT show distant items (use specific patterns)
    refute_output_contains "○ 1. End 1"
    refute_output_contains "○ 2. End 2"
    refute_output_contains "○ 3. End 3"
}