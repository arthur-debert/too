#!/usr/bin/env bats

# Test suite for plain text addressing feature (Issue #180)
# Tests: completing, editing, and moving todos by their text content instead of numeric indices

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

@test "text addressing: complete todo by exact text match" {
    too init >/dev/null 2>&1
    
    # Create todos
    too add "Buy groceries" >/dev/null 2>&1
    too add "Walk the dog" >/dev/null 2>&1
    too add "Write documentation" >/dev/null 2>&1
    
    # Complete by exact text
    run too complete "Walk the dog"
    [ "$status" -eq 0 ]
    
    # Verify it's marked as done
    run too list --all
    assert_output_contains "● c1. Walk the dog"
}

@test "text addressing: edit todo by partial text match" {
    too init >/dev/null 2>&1
    
    # Create todos
    too add "Buy milk" >/dev/null 2>&1
    too add "Buy bread" >/dev/null 2>&1
    too add "Walk the dog" >/dev/null 2>&1
    
    # Edit by partial match (unique)
    run too edit "dog" "Walk the cat"
    [ "$status" -eq 0 ]
    
    # Verify the edit
    run too list
    assert_output_contains "Walk the cat"
    refute_output_contains "Walk the dog"
}

@test "text addressing: ambiguous text shows helpful error" {
    too init >/dev/null 2>&1
    
    # Create ambiguous todos
    too add "Write tests" >/dev/null 2>&1
    too add "Write documentation" >/dev/null 2>&1
    too add "Write blog post" >/dev/null 2>&1
    
    # Try to complete with ambiguous text
    run too complete "Write"
    [ "$status" -ne 0 ]
    
    # Should show multiple matches
    assert_output_contains "Multiple todos found"
    assert_output_contains "1: Write tests"
    assert_output_contains "2: Write documentation"
    assert_output_contains "3: Write blog post"
    assert_output_contains "Please be more specific"
}

@test "text addressing: exact match resolves ambiguity" {
    too init >/dev/null 2>&1
    
    # Create todos where one is an exact match
    too add "Test" >/dev/null 2>&1
    too add "Test the feature" >/dev/null 2>&1
    too add "Testing framework" >/dev/null 2>&1
    
    # Complete by exact match
    run too complete "Test"
    [ "$status" -eq 0 ]
    
    # Verify only the exact match was completed
    run too list --all
    assert_output_contains "● c1. Test"
    assert_output_contains "○ 1. Test the feature"
    assert_output_contains "○ 2. Testing framework"
}

@test "text addressing: case-insensitive matching" {
    too init >/dev/null 2>&1
    
    # Create todo with mixed case
    too add "Buy MILK and Bread" >/dev/null 2>&1
    too add "Walk the Dog" >/dev/null 2>&1
    
    # Complete with different case
    run too complete "buy milk"
    [ "$status" -eq 0 ]
    
    # Edit with different case
    run too edit "walk the DOG" "Feed the cat"
    [ "$status" -eq 0 ]
    
    # Verify both operations worked
    run too list --all
    assert_output_contains "● c1. Buy MILK and Bread"
    assert_output_contains "○ 1. Feed the cat"
}

@test "text addressing: works with nested todos" {
    too init >/dev/null 2>&1
    
    # Create nested structure
    too add "Shopping" >/dev/null 2>&1
    too add --to 1 "Buy milk" >/dev/null 2>&1
    too add --to 1 "Buy bread" >/dev/null 2>&1
    too add "Chores" >/dev/null 2>&1
    too add --to 2 "Walk the dog" >/dev/null 2>&1
    too add --to 2 "Clean the house" >/dev/null 2>&1
    
    # Complete nested todo by text
    run too complete "Buy bread"
    [ "$status" -eq 0 ]
    
    # Verify completion worked
    run too list --all
    assert_output_contains "● 1.c1. Buy bread"
    
    # Also complete the parent shopping task
    run too complete "Shopping"
    [ "$status" -eq 0 ]
    
    # Verify parent is completed
    run too list --all
    assert_output_contains "◐ c1. Shopping"
}

@test "text addressing: numeric positions still work" {
    too init >/dev/null 2>&1
    
    # Create todos
    too add "First todo" >/dev/null 2>&1
    too add "Second todo" >/dev/null 2>&1
    too add "99" >/dev/null 2>&1  # Todo with numeric text
    
    # Complete by position
    run too complete 2
    [ "$status" -eq 0 ]
    
    # Complete todo with numeric text by position (now at position 2 after first completion)
    run too complete 2
    [ "$status" -eq 0 ]
    
    run too list --all
    assert_output_contains "○ 1. First todo"
    assert_output_contains "● c1. Second todo"
    assert_output_contains "● c2. 99"
}

@test "text addressing: no match error" {
    too init >/dev/null 2>&1
    
    # Create todos
    too add "Buy milk" >/dev/null 2>&1
    too add "Walk the dog" >/dev/null 2>&1
    
    # Try to complete non-existent todo
    run too complete "Feed the cat"
    [ "$status" -ne 0 ]
    assert_output_contains "no todo found matching 'Feed the cat'"
}

@test "text addressing: works with multiple commands" {
    too init >/dev/null 2>&1
    
    # Create todos
    too add "Task one" >/dev/null 2>&1
    too add "Task two" >/dev/null 2>&1
    too add "Task three" >/dev/null 2>&1
    
    # Complete multiple by text (if supported)
    run too complete "Task one" "Task three"
    [ "$status" -eq 0 ]
    
    run too list --all
    assert_output_contains "● c1. Task one"
    assert_output_contains "○ 1. Task two"
    assert_output_contains "● c2. Task three"
}

@test "text addressing: special characters in text" {
    too init >/dev/null 2>&1
    
    # Create todos with special characters
    too add "Buy milk @ store" >/dev/null 2>&1
    too add "Review PR #123" >/dev/null 2>&1
    too add "Fix bug: memory leak" >/dev/null 2>&1
    
    # Complete by text with special chars
    run too complete "PR #123"
    [ "$status" -eq 0 ]
    
    run too edit "bug: memory" "Fix bug: null pointer"
    [ "$status" -eq 0 ]
    
    run too list --all
    assert_output_contains "● c1. Review PR #123"
    assert_output_contains "Fix bug: null pointer"
}