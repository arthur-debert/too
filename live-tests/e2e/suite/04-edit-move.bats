#!/usr/bin/env bats

# Test suite for edit and move operations
# Tests: edit text, move to different parent, reorder within parent

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

@test "edit todo text" {
    # Create a test script that edits a todo
    cat > "$TEST_DIR/edit_text.sh" << 'EOF'
#!/bin/zsh
# Create a todo and edit it
too add "Original text" --format "${TOO_FORMAT}"
too edit 1 "Updated text" --format "${TOO_FORMAT}"

too list --format "${TOO_FORMAT}"
EOF
    chmod +x "$TEST_DIR/edit_text.sh"
    
    # Run the test script with JSON format
    output=$("$LIVE_TESTS_DIR/run" --format json "$TEST_DIR/edit_text.sh")
    
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
    
    # Verify the text was updated
    total_count=$(get_total_count "$json_output")
    [ "$total_count" -eq 1 ]
    
    # Should have updated text, not original
    todo_exists "$json_output" "Updated text"
    ! todo_exists "$json_output" "Original text"
    
    # Clean up
    rm -f "$TEST_DIR/edit_text.sh"
}

@test "move todo to different parent" {
    # Create a test script that moves a todo between parents
    cat > "$TEST_DIR/move_parent.sh" << 'EOF'
#!/bin/zsh
# Create structure
too add "Parent A" --format "${TOO_FORMAT}"
too add "Parent B" --format "${TOO_FORMAT}"
too add --to 1 "Child of A" --format "${TOO_FORMAT}"

# Move child from Parent A to Parent B
too move 1.1 2 --format "${TOO_FORMAT}"

too list --format "${TOO_FORMAT}"
EOF
    chmod +x "$TEST_DIR/move_parent.sh"
    
    # Run the test script with JSON format
    output=$("$LIVE_TESTS_DIR/run" --format json "$TEST_DIR/move_parent.sh")
    
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
    
    # Verify the move happened
    validate_parent_child "$json_output" "Parent B" "Child of A"
    
    # Verify it's no longer under Parent A
    parent_a=$(get_todo_by_text "$json_output" "Parent A")
    parent_a_uid=$(echo "$parent_a" | jq -r '.uid')
    
    child=$(get_todo_by_text "$json_output" "Child of A")
    child_parent=$(echo "$child" | jq -r '.parentId')
    
    [ "$child_parent" != "$parent_a_uid" ]
    
    # Clean up
    rm -f "$TEST_DIR/move_parent.sh"
}

@test "reorder todos within same parent" {
    # Create a test script that reorders siblings
    cat > "$TEST_DIR/reorder_siblings.sh" << 'EOF'
#!/bin/zsh
# Create parent with multiple children
too add "Parent" --format "${TOO_FORMAT}"
too add --to 1 "Child 1" --format "${TOO_FORMAT}"
too add --to 1 "Child 2" --format "${TOO_FORMAT}"
too add --to 1 "Child 3" --format "${TOO_FORMAT}"

# List before reorder
too list --format "${TOO_FORMAT}"
echo "SEPARATOR"

# Move Child 3 to position of Child 1 (reorder)
too move 1.3 1.1 --format "${TOO_FORMAT}"

# List after reorder
too list --format "${TOO_FORMAT}"
EOF
    chmod +x "$TEST_DIR/reorder_siblings.sh"
    
    # Run the test script with JSON format
    output=$("$LIVE_TESTS_DIR/run" --format json "$TEST_DIR/reorder_siblings.sh")
    
    # Extract the final list command JSON output (after SEPARATOR)
    json_output=$(echo "$output" | awk '/SEPARATOR/{p=1;next} p' | awk '
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
    
    # All items should still exist
    total_count=$(get_total_count "$json_output")
    [ "$total_count" -eq 4 ]
    
    todo_exists "$json_output" "Parent"
    todo_exists "$json_output" "Child 1"
    todo_exists "$json_output" "Child 2"
    todo_exists "$json_output" "Child 3"
    
    # Clean up
    rm -f "$TEST_DIR/reorder_siblings.sh"
}