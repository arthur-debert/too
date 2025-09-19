#!/usr/bin/env bats

# Test global vs local (project) scope functionality

setup() {
    # Get the directory containing this test file
    TEST_DIR="$( cd "$( dirname "$BATS_TEST_FILENAME" )" && pwd )"
    E2E_DIR="$(dirname "$TEST_DIR")"
    LIVE_TESTS_DIR="$(dirname "$E2E_DIR")"
    PROJECT_ROOT="$(dirname "$LIVE_TESTS_DIR")"
    
    # Export paths for use in tests
    export TEST_DIR
    export E2E_DIR
    export LIVE_TESTS_DIR
    export PROJECT_ROOT
    
    # Create a temporary directory for this test
    export TEST_TEMP_DIR="$(mktemp -d)"
    
    # Build the too binary if needed
    if [[ ! -f "$PROJECT_ROOT/bin/too" ]]; then
        (cd "$PROJECT_ROOT" && go build -o bin/too ./cmd/too)
    fi
    
    # Use the actual binary path
    export TOO_BIN="$PROJECT_ROOT/bin/too"
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
    if ! echo "$output" | grep -q "$expected"; then
        echo "Expected output to contain: $expected"
        echo "But got: $output"
        return 1
    fi
}

# Helper function to check output does not contain text
assert_output_not_contains() {
    local unexpected="$1"
    if echo "$output" | grep -q "$unexpected"; then
        echo "Expected output NOT to contain: $unexpected"
        echo "But got: $output"
        return 1
    fi
}

@test "scope: default to global when not in git repo" {
    # Create a non-git directory
    mkdir -p "$TEST_TEMP_DIR/non-git"
    cd "$TEST_TEMP_DIR/non-git"
    
    # Clear any environment variables that might interfere
    unset TODO_DB_PATH
    
    # Add a todo (should use global scope)
    "$TOO_BIN" add "Global todo 1"
    
    # The datapath should be in XDG_DATA_HOME or ~/.local/share
    run "$TOO_BIN" datapath
    assert_output_contains "/.local/share/too/todos.json"
}

@test "scope: use project scope in git repo" {
    # Create a git repository
    mkdir -p "$TEST_TEMP_DIR/git-repo"
    cd "$TEST_TEMP_DIR/git-repo"
    git init
    
    # Clear any environment variables that might interfere
    unset TODO_DB_PATH
    
    # Add a todo (should use project scope)
    "$TOO_BIN" add "Project todo 1"
    
    # The datapath should be in the git repo
    run "$TOO_BIN" datapath
    [ "$output" = "$TEST_TEMP_DIR/git-repo/.todos.json" ]
    
    # Check .gitignore was updated
    [ -f .gitignore ]
    grep -q ".todos.json" .gitignore
}

@test "scope: --global flag forces global scope in git repo" {
    # Create a git repository
    mkdir -p "$TEST_TEMP_DIR/git-repo-global"
    cd "$TEST_TEMP_DIR/git-repo-global"
    git init
    
    # Clear any environment variables that might interfere
    unset TODO_DB_PATH
    
    # Add a todo with --global flag
    "$TOO_BIN" --global add "Forced global todo"
    
    # The datapath should be global, not project
    run "$TOO_BIN" --global datapath
    assert_output_contains "/.local/share/too/todos.json"
    
    # Project .todos.json should NOT exist
    [ ! -f ".todos.json" ]
}

@test "scope: todos are isolated between scopes" {
    # Create a git repo for project scope
    mkdir -p "$TEST_TEMP_DIR/scope-isolation"
    cd "$TEST_TEMP_DIR/scope-isolation"
    git init
    
    # Clear any environment variables that might interfere
    unset TODO_DB_PATH
    
    # Add project todo
    "$TOO_BIN" add "Project-specific todo"
    
    # Add global todo
    "$TOO_BIN" --global add "Global todo"
    
    # List project todos (should not show global todo)
    run "$TOO_BIN" list
    assert_output_contains "Project-specific todo"
    assert_output_not_contains "Global todo"
    
    # List global todos (should not show project todo)
    run "$TOO_BIN" --global list
    assert_output_contains "Global todo"
    assert_output_not_contains "Project-specific todo"
}

@test "scope: subdirectory finds parent git repo" {
    # Create git repo
    mkdir -p "$TEST_TEMP_DIR/parent-search"
    cd "$TEST_TEMP_DIR/parent-search"
    git init
    
    # Clear any environment variables that might interfere
    unset TODO_DB_PATH
    
    # Create deep subdirectory
    mkdir -p src/pkg/utils
    cd src/pkg/utils
    
    # Add todo from subdirectory
    "$TOO_BIN" add "Subdirectory todo"
    
    # Check .todos.json is in git root
    [ -f "$TEST_TEMP_DIR/parent-search/.todos.json" ]
    [ ! -f ".todos.json" ]
    
    # List should still work
    run "$TOO_BIN" list
    assert_output_contains "Subdirectory todo"
}