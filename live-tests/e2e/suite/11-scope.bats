#!/usr/bin/env bats

# Test global vs local (project) scope functionality

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
}

@test "scope: default to global when not in git repo" {
    # Create test script for non-git directory
    cat > "$TEST_DIR/scope_global.sh" << 'EOF'
#!/bin/zsh
# Create a non-git directory
TEST_DIR=$(mktemp -d)
cd "$TEST_DIR"

# Add a todo (should use global scope)
too add "Global todo 1"

# The datapath should be in XDG_DATA_HOME or ~/.local/share
too datapath

# Clean up
rm -rf "$TEST_DIR"
EOF
    chmod +x "$TEST_DIR/scope_global.sh"
    
    # Run the test script
    output=$("$LIVE_TESTS_DIR/run" "$TEST_DIR/scope_global.sh")
    
    # Check that it uses global path
    echo "$output" | grep -q "/.local/share/too/todos.json" || {
        echo "Expected global path, but got: $output"
        return 1
    }
}

@test "scope: use project scope in git repo" {
    # Create test script for git repo
    cat > "$TEST_DIR/scope_project.sh" << 'EOF'
#!/bin/zsh
# Create a git repository
TEST_DIR=$(mktemp -d)
cd "$TEST_DIR"
git init

# Add a todo (should use project scope)
too add "Project todo 1"

# The datapath should be in the git repo
too datapath

# Check .gitignore
cat .gitignore 2>/dev/null | grep -q ".todos.json" && echo "GITIGNORE_OK"

# Clean up
rm -rf "$TEST_DIR"
EOF
    chmod +x "$TEST_DIR/scope_project.sh"
    
    # Run the test script
    output=$("$LIVE_TESTS_DIR/run" "$TEST_DIR/scope_project.sh")
    
    # Check that it uses project path
    echo "$output" | grep -q "/.todos.json" || {
        echo "Expected project path ending with .todos.json"
        return 1
    }
    
    # Check gitignore was updated
    echo "$output" | grep -q "GITIGNORE_OK" || {
        echo ".gitignore was not updated properly"
        return 1
    }
}

@test "scope: --global flag forces global scope in git repo" {
    # Create test script
    cat > "$TEST_DIR/scope_force_global.sh" << 'EOF'
#!/bin/zsh
# Create a git repository
TEST_DIR=$(mktemp -d)
cd "$TEST_DIR"
git init

# Add a todo with --global flag
too --global add "Forced global todo"

# The datapath should be global, not project
too --global datapath

# Project .todos.json should NOT exist
[ ! -f ".todos.json" ] && echo "NO_PROJECT_FILE"

# Clean up
rm -rf "$TEST_DIR"
EOF
    chmod +x "$TEST_DIR/scope_force_global.sh"
    
    # Run the test script
    output=$("$LIVE_TESTS_DIR/run" "$TEST_DIR/scope_force_global.sh")
    
    # Check that it uses global path
    echo "$output" | grep -q "/.local/share/too/todos.json" || {
        echo "Expected global path, but got: $output"
        return 1
    }
    
    # Check no project file was created
    echo "$output" | grep -q "NO_PROJECT_FILE" || {
        echo "Project .todos.json should not exist"
        return 1
    }
}

@test "scope: todos are isolated between scopes" {
    # Create test script
    cat > "$TEST_DIR/scope_isolation.sh" << 'EOF'
#!/bin/zsh
# Create a git repo for project scope
TEST_DIR=$(mktemp -d)
cd "$TEST_DIR"
git init

# Add project todo
too add "Project-specific todo"

# Add global todo
too --global add "Global todo"

# List project todos (should not show global todo)
echo "=== PROJECT TODOS ==="
too list

# List global todos (should not show project todo)
echo "=== GLOBAL TODOS ==="
too --global list

# Clean up
rm -rf "$TEST_DIR"
EOF
    chmod +x "$TEST_DIR/scope_isolation.sh"
    
    # Run the test script
    output=$("$LIVE_TESTS_DIR/run" "$TEST_DIR/scope_isolation.sh")
    
    # Extract project todos section
    project_section=$(echo "$output" | sed -n '/=== PROJECT TODOS ===/,/=== GLOBAL TODOS ===/p')
    
    # Check project section has project todo but not global
    echo "$project_section" | grep -q "Project-specific todo" || {
        echo "Project todos should contain 'Project-specific todo'"
        return 1
    }
    echo "$project_section" | grep -q "Global todo" && {
        echo "Project todos should NOT contain 'Global todo'"
        return 1
    }
    
    # Extract global todos section  
    global_section=$(echo "$output" | sed -n '/=== GLOBAL TODOS ===/,$p')
    
    # Check global section has global todo but not project
    echo "$global_section" | grep -q "Global todo" || {
        echo "Global todos should contain 'Global todo'"
        return 1
    }
    echo "$global_section" | grep -q "Project-specific todo" && {
        echo "Global todos should NOT contain 'Project-specific todo'"
        return 1
    }
}

@test "scope: subdirectory finds parent git repo" {
    # Create test script
    cat > "$TEST_DIR/scope_subdirectory.sh" << 'EOF'
#!/bin/zsh
# Create git repo
TEST_DIR=$(mktemp -d)
cd "$TEST_DIR"
git init

# Create deep subdirectory
mkdir -p src/pkg/utils
cd src/pkg/utils

# Add todo from subdirectory
too add "Subdirectory todo"

# Check .todos.json is in git root
[ -f "../../../.todos.json" ] && echo "TODOS_IN_ROOT"
[ ! -f ".todos.json" ] && echo "NO_LOCAL_TODOS"

# List should still work
too list | grep -q "Subdirectory todo" && echo "LIST_WORKS"

# Clean up
cd /
rm -rf "$TEST_DIR"
EOF
    chmod +x "$TEST_DIR/scope_subdirectory.sh"
    
    # Run the test script
    output=$("$LIVE_TESTS_DIR/run" "$TEST_DIR/scope_subdirectory.sh")
    
    # Check all conditions
    echo "$output" | grep -q "TODOS_IN_ROOT" || {
        echo ".todos.json should be in git root"
        return 1
    }
    echo "$output" | grep -q "NO_LOCAL_TODOS" || {
        echo ".todos.json should not be in subdirectory"
        return 1
    }
    echo "$output" | grep -q "LIST_WORKS" || {
        echo "List command should work from subdirectory"
        return 1
    }
}