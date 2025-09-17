#!/usr/bin/env bats

# Test suite for creating todos with special text content
# Tests: todos with newlines, special characters, long text

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

@test "add todo with internal line breaks" {
    # Create a test script that adds a todo with line breaks
    cat > "$TEST_DIR/todo_with_breaks.sh" << 'EOF'
#!/bin/zsh
# Add a todo with escaped newlines
too add "This is a todo\\nwith multiple\\nlines of text" --format "${TOO_FORMAT}"

too list --format "${TOO_FORMAT}"
EOF
    chmod +x "$TEST_DIR/todo_with_breaks.sh"
    
    # Run the test script with JSON format
    output=$("$LIVE_TESTS_DIR/run" --format json "$TEST_DIR/todo_with_breaks.sh")
    
    # Extract the list command JSON output
    json_output=$(echo "$output" | awk '
        BEGIN { in_json=0; current_json="" }
        /^{$/ { in_json=1; current_json=$0"\n"; next }
        in_json==1 { current_json=current_json$0"\n" }
        /^}$/ { 
            in_json=0; 
            if (current_json ~ /"Command": "list"/) {
                list_output = current_json
            }
            current_json=""
        }
        END { print list_output }')
    
    # Should create one todo with the full text
    total_count=$(get_total_count "$json_output")
    [ "$total_count" -eq 1 ]
    
    # Verify the todo contains the expected text pattern
    todo_text=$(echo "$json_output" | jq -r '.AllTodos[0].text')
    [[ "$todo_text" == *"multiple"* ]]
    [[ "$todo_text" == *"lines"* ]]
    
    # Clean up
    rm -f "$TEST_DIR/todo_with_breaks.sh"
}

@test "add multiple todos in single command" {
    # Create a test script that adds multiple todos sequentially
    cat > "$TEST_DIR/multiple_todos.sh" << 'EOF'
#!/bin/zsh
# Add multiple todos in one go
too add "Task 1" --format "${TOO_FORMAT}"
too add "Task 2" --format "${TOO_FORMAT}"
too add "Task 3" --format "${TOO_FORMAT}"

too list --format "${TOO_FORMAT}"
EOF
    chmod +x "$TEST_DIR/multiple_todos.sh"
    
    # Run the test script with JSON format
    output=$("$LIVE_TESTS_DIR/run" --format json "$TEST_DIR/multiple_todos.sh")
    
    # Extract the list command JSON output
    json_output=$(echo "$output" | awk '
        BEGIN { in_json=0; current_json="" }
        /^{$/ { in_json=1; current_json=$0"\n"; next }
        in_json==1 { current_json=current_json$0"\n" }
        /^}$/ { 
            in_json=0; 
            if (current_json ~ /"Command": "list"/) {
                list_output = current_json
            }
            current_json=""
        }
        END { print list_output }')
    
    # Should create three separate todos
    total_count=$(get_total_count "$json_output")
    [ "$total_count" -eq 3 ]
    
    # Verify each todo was created
    todo_exists "$json_output" "Task 1"
    todo_exists "$json_output" "Task 2"
    todo_exists "$json_output" "Task 3"
    
    # Clean up
    rm -f "$TEST_DIR/multiple_todos.sh"
}

@test "add todo with very long text" {
    # Create a test script with a long todo
    cat > "$TEST_DIR/long_todo.sh" << 'EOF'
#!/bin/zsh
# Add a todo with very long text
too add "This is a very long todo item that contains a lot of text to test how the system handles extended descriptions. It should handle this gracefully without truncation or issues. Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua." --format "${TOO_FORMAT}"

too list --format "${TOO_FORMAT}"
EOF
    chmod +x "$TEST_DIR/long_todo.sh"
    
    # Run the test script with JSON format
    output=$("$LIVE_TESTS_DIR/run" --format json "$TEST_DIR/long_todo.sh")
    
    # Extract the list command JSON output
    json_output=$(echo "$output" | awk '
        BEGIN { in_json=0; current_json="" }
        /^{$/ { in_json=1; current_json=$0"\n"; next }
        in_json==1 { current_json=current_json$0"\n" }
        /^}$/ { 
            in_json=0; 
            if (current_json ~ /"Command": "list"/) {
                list_output = current_json
            }
            current_json=""
        }
        END { print list_output }')
    
    # Should create one todo
    total_count=$(get_total_count "$json_output")
    [ "$total_count" -eq 1 ]
    
    # Verify the full text is preserved
    todo_text=$(echo "$json_output" | jq -r '.AllTodos[0].text')
    [[ "$todo_text" == *"Lorem ipsum"* ]]
    [[ "$todo_text" == *"magna aliqua"* ]]
    
    # Clean up
    rm -f "$TEST_DIR/long_todo.sh"
}

@test "add nested todos using 'to' keyword" {
    # Create a test script that creates nested structure
    cat > "$TEST_DIR/nested_todos.sh" << 'EOF'
#!/bin/zsh
# Create nested todo structure
too add "Project Setup" --format "${TOO_FORMAT}"
too add --to 1 "Install dependencies" --format "${TOO_FORMAT}"
too add --to 1 "Configure environment" --format "${TOO_FORMAT}"
too add --to 1 "Create database" --format "${TOO_FORMAT}"

too list --format "${TOO_FORMAT}"
EOF
    chmod +x "$TEST_DIR/nested_todos.sh"
    
    # Run the test script with JSON format
    output=$("$LIVE_TESTS_DIR/run" --format json "$TEST_DIR/nested_todos.sh")
    
    # Extract the list command JSON output
    json_output=$(echo "$output" | awk '
        BEGIN { in_json=0; current_json="" }
        /^{$/ { in_json=1; current_json=$0"\n"; next }
        in_json==1 { current_json=current_json$0"\n" }
        /^}$/ { 
            in_json=0; 
            if (current_json ~ /"Command": "list"/) {
                list_output = current_json
            }
            current_json=""
        }
        END { print list_output }')
    
    # Should create 4 todos total
    total_count=$(get_total_count "$json_output")
    [ "$total_count" -eq 4 ]
    
    # Verify parent-child relationships
    validate_parent_child "$json_output" "Project Setup" "Install dependencies"
    validate_parent_child "$json_output" "Project Setup" "Configure environment"
    validate_parent_child "$json_output" "Project Setup" "Create database"
    
    # Clean up
    rm -f "$TEST_DIR/nested_todos.sh"
}