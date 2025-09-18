#!/usr/bin/env bats

# Test suite for naked execution feature
# Tests: naked list, naked add, flags handling

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

@test "naked execution: no args defaults to list" {
    too init >/dev/null 2>&1
    too add "First task"
    too add "Second task"
    
    run too
    [ "$status" -eq 0 ]
    assert_output_contains "○ 1. First task"
    assert_output_contains "○ 2. Second task"
}

@test "naked execution: single word creates todo" {
    too init >/dev/null 2>&1
    
    run too Groceries
    [ "$status" -eq 0 ]
    assert_output_contains "○ 1. Groceries"
}

@test "naked execution: multiple words creates todo" {
    too init >/dev/null 2>&1
    
    run too Buy milk and bread
    [ "$status" -eq 0 ]
    assert_output_contains "○ 1. Buy milk and bread"
}

@test "naked execution: with parent flag" {
    too init >/dev/null 2>&1
    too add "Shopping"
    
    run too --to 1 Milk
    [ "$status" -eq 0 ]
    assert_output_contains "○ 1. Shopping"
    assert_output_contains "  ○ 1.1. Milk"
}

@test "naked execution: multiple adds with parent" {
    too init >/dev/null 2>&1
    too add "Groceries"
    
    too --to 1 Apples
    too --to 1 Bananas
    run too --to 1 Oranges
    
    [ "$status" -eq 0 ]
    assert_output_contains "○ 1. Groceries"
    assert_output_contains "  ○ 1.1. Apples"
    assert_output_contains "  ○ 1.2. Bananas"
    assert_output_contains "  ○ 1.3. Oranges"
}

@test "naked execution: regular commands still work" {
    too init >/dev/null 2>&1
    
    # Test regular add
    run too add "Regular add task"
    [ "$status" -eq 0 ]
    assert_output_contains "○ 1. Regular add task"
    
    # Test regular list
    run too list
    [ "$status" -eq 0 ]
    assert_output_contains "○ 1. Regular add task"
}

@test "naked execution: with global flags" {
    too init >/dev/null 2>&1
    
    # Add with verbose flag
    run too -v Debug message test
    [ "$status" -eq 0 ]
    assert_output_contains "○ 1. Debug message test"
}

@test "naked execution: help flag shows help" {
    run too --help
    [ "$status" -eq 0 ]
    assert_output_contains "Fast project-aware, nested, command-line todo list"
    assert_output_contains "USAGE:"
}

@test "naked execution: version flag shows version" {
    run too --version
    [ "$status" -eq 0 ]
    assert_output_contains "too version"
}

@test "naked execution: with format flag" {
    too init >/dev/null 2>&1
    
    run too --format json Task item
    [ "$status" -eq 0 ]
    # Should output JSON format
    assert_output_contains "Task item"
    assert_output_contains "AffectedTodos"
}

@test "naked execution: empty list with format" {
    too init >/dev/null 2>&1
    
    run too --format json
    [ "$status" -eq 0 ]
    # Should output JSON with list command result
    assert_output_contains "list"
    assert_output_contains "TotalCount"
}

@test "naked execution: preserves quotes in text" {
    too init >/dev/null 2>&1
    
    run too "Buy \"special\" bread"
    [ "$status" -eq 0 ]
    assert_output_contains 'Buy "special" bread'
}

@test "naked execution: handles special characters" {
    too init >/dev/null 2>&1
    
    run too "Task with & symbols | pipes > redirects"
    [ "$status" -eq 0 ]
    assert_output_contains "Task with & symbols | pipes > redirects"
}

@test "naked execution: unknown command creates todo" {
    too init >/dev/null 2>&1
    
    # What looks like a command but isn't should create a todo
    run too notacommand some text
    [ "$status" -eq 0 ]
    assert_output_contains "○ 1. notacommand some text"
}

@test "naked execution: flags without args shows list" {
    too init >/dev/null 2>&1
    too add "Task one"
    too complete 1
    too add "Task two"
    
    # Just flags, no args - should list
    run too list --done
    [ "$status" -eq 0 ]
    assert_output_contains "● c1. Task one"
    refute_output_contains "○ 2. Task two"
}

@test "naked execution: complex parent chain" {
    too init >/dev/null 2>&1
    too add "Project"
    too --to 1 "Module A"
    too --to 1.1 "Component X"
    
    run too --to 1.1.1 "Detail item"
    [ "$status" -eq 0 ]
    assert_output_contains "○ 1. Project"
    assert_output_contains "  ○ 1.1. Module A"
    assert_output_contains "    ○ 1.1.1. Component X"
    assert_output_contains "      ○ 1.1.1.1. Detail item"
}
