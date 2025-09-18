#!/usr/bin/env bats

# Sample e2e test demonstrating JSON parsing and validation
# This shows how to write comprehensive tests using the utility functions

setup() {
    # Get the directory containing this test file
    TEST_DIR="$( cd "$( dirname "$BATS_TEST_FILENAME" )" && pwd )"
    LIVE_TESTS_DIR="$(dirname "$TEST_DIR")"
    PROJECT_ROOT="$(dirname "$LIVE_TESTS_DIR")"
    
    # Load utility functions
    source "$TEST_DIR/utils/parse-json.sh"
    source "$TEST_DIR/utils/generate-sample.sh"
    
    # Export paths for use in tests
    export TEST_DIR
    export LIVE_TESTS_DIR 
    export PROJECT_ROOT
}

@test "hierarchical todo structure validation" {
    # Generate a test script with hierarchical structure
    generate_basic_operations_script "$TEST_DIR/hierarchy_test.sh"
    
    # Run the test script with JSON format
    output=$("$LIVE_TESTS_DIR/run" --format json "$TEST_DIR/hierarchy_test.sh")
    
    # Extract the final list command JSON output
    # Find the last JSON block that contains "Command": "list"
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
    
    # Validate total counts
    total_count=$(get_total_count "$json_output")
    done_count=$(get_done_count "$json_output")
    
    [ "$total_count" -eq 7 ]  # Should have 7 todos total
    [ "$done_count" -eq 2 ]   # Should have 2 completed todos
    
    # Validate specific todos exist
    todo_exists "$json_output" "Buy groceries"
    todo_exists "$json_output" "Apples" 
    todo_exists "$json_output" "Book hotel"
    
    # Validate parent-child relationships
    validate_parent_child "$json_output" "Buy groceries" "Bread"
    validate_parent_child "$json_output" "Plan weekend trip" "Book hotel"
    validate_parent_child "$json_output" "Plan weekend trip" "Check weather"
    
    # Validate completion status
    apples_todo=$(get_todo_by_text "$json_output" "Apples")
    apples_status=$(echo "$apples_todo" | jq -r '.statuses.completion')
    [ "$apples_status" = "done" ]
    
    # Validate pending status  
    bread_todo=$(get_todo_by_text "$json_output" "Bread")
    bread_status=$(echo "$bread_todo" | jq -r '.statuses.completion')
    [ "$bread_status" = "pending" ]
    
    # Count todos by status
    pending_count=$(count_todos_by_status "$json_output" "pending")
    completed_count=$(count_todos_by_status "$json_output" "done")
    
    [ "$pending_count" -eq 5 ]
    [ "$completed_count" -eq 2 ]
    
    # Clean up
    rm -f "$TEST_DIR/hierarchy_test.sh"
    
    echo "✅ Hierarchical structure validation passed"
}

@test "flat list operations" {
    # Create a simple flat list test script
    cat > "$TEST_DIR/flat_test.sh" << 'EOF'
#!/bin/zsh
too add "Task A"
too add "Task B"
too add "Task C"
too complete 2
too list
EOF
    chmod +x "$TEST_DIR/flat_test.sh"
    
    # Run the test script
    output=$("$LIVE_TESTS_DIR/run" --format json "$TEST_DIR/flat_test.sh")
    
    # Extract the final list command JSON output  
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
    
    # Validate structure
    total_count=$(get_total_count "$json_output")
    [ "$total_count" -eq 3 ]
    
    # Check that all todos are root level (no parent)
    root_todos=$(get_todos_by_parent "$json_output" "")
    root_count=$(echo "$root_todos" | jq -s 'length')
    [ "$root_count" -eq 3 ]
    
    # Validate specific todo status
    task_b_todo=$(get_todo_by_text "$json_output" "Task B")
    task_b_status=$(echo "$task_b_todo" | jq -r '.statuses.completion')
    [ "$task_b_status" = "done" ]
    
    # Clean up
    rm -f "$TEST_DIR/flat_test.sh"
    
    echo "✅ Flat list operations passed"
}

@test "edge cases and special characters" {
    # Generate edge cases test script
    generate_edge_cases_script "$TEST_DIR/edge_test.sh"
    
    # Run the test script
    output=$("$LIVE_TESTS_DIR/run" --format json "$TEST_DIR/edge_test.sh")
    
    # Extract the final list command JSON output  
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
    
    # Should handle quotes properly
    todo_exists "$json_output" "Todo with \"quotes\" and 'apostrophes'"
    
    # Should handle unicode
    todo_exists "$json_output" "Todo with unicode: 🚀 🎯 ✅"
    
    # Should handle long text
    long_text_todo=$(get_todo_by_text "$json_output" "This is a very long todo item that has a lot of text to test how the system handles longer descriptions and whether everything works correctly with extended content")
    [ -n "$long_text_todo" ]
    
    # Should handle nested hierarchy
    validate_parent_child "$json_output" "Level 1" "Level 2"
    validate_parent_child "$json_output" "Level 2" "Level 3"
    
    # Clean up
    rm -f "$TEST_DIR/edge_test.sh"
    
    echo "✅ Edge cases validation passed"
}