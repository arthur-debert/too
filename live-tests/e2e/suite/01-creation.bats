#!/usr/bin/env bats

# Test suite for basic creation operations
# Tests: add at root level, add with parent, create grandchildren, verify hierarchy

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

@test "create item at top level" {
    # Create a test script that adds a single root-level item
    cat > "$TEST_DIR/create_root.sh" << 'EOF'
#!/bin/zsh
too add "First todo"
too list
EOF
    chmod +x "$TEST_DIR/create_root.sh"
    
    # Run the test script with JSON format
    output=$("$LIVE_TESTS_DIR/run" --format json "$TEST_DIR/create_root.sh")
    
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
    
    # Validate the todo was created
    total_count=$(get_total_count "$json_output")
    [ "$total_count" -eq 1 ]
    
    # Verify it exists and has no parent (root level)
    todo=$(get_todo_by_text "$json_output" "First todo")
    parent_id=$(echo "$todo" | jq -r '.parentId')
    [ "$parent_id" = "" ]
    
    # Clean up
    rm -f "$TEST_DIR/create_root.sh"
}

@test "create two items at top level" {
    # Create a test script that adds two root-level items
    cat > "$TEST_DIR/create_two_root.sh" << 'EOF'
#!/bin/zsh
too add "First todo"
too add "Second todo"
too list
EOF
    chmod +x "$TEST_DIR/create_two_root.sh"
    
    # Run the test script with JSON format
    output=$("$LIVE_TESTS_DIR/run" --format json "$TEST_DIR/create_two_root.sh")
    
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
    
    # Validate both todos were created
    total_count=$(get_total_count "$json_output")
    [ "$total_count" -eq 2 ]
    
    # Verify both exist and have no parent (root level)
    todo_exists "$json_output" "First todo"
    todo_exists "$json_output" "Second todo"
    
    # Get todos and verify they're root level
    root_todos=$(get_todos_by_parent "$json_output" "")
    root_count=$(echo "$root_todos" | grep -c '"text"')
    [ "$root_count" -eq 2 ]
    
    # Clean up
    rm -f "$TEST_DIR/create_two_root.sh"
}

@test "create children for each parent" {
    # Create a test script that adds parents and children
    cat > "$TEST_DIR/create_with_children.sh" << 'EOF'
#!/bin/zsh
too add "Groceries"
too add "Pack"
too add --to 1 "Bread"
too add --to 1 "Milk"
too add --to 1 "Pancakes"
too add --to 2 "Camera"
too add --to 2 "Clothes"
too add --to 2 "Passport"
too list
EOF
    chmod +x "$TEST_DIR/create_with_children.sh"
    
    # Run the test script with JSON format
    output=$("$LIVE_TESTS_DIR/run" --format json "$TEST_DIR/create_with_children.sh")
    
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
    
    # Validate total count
    total_count=$(get_total_count "$json_output")
    [ "$total_count" -eq 8 ]
    
    # Verify parent-child relationships
    validate_parent_child "$json_output" "Groceries" "Bread"
    validate_parent_child "$json_output" "Groceries" "Milk"
    validate_parent_child "$json_output" "Groceries" "Pancakes"
    validate_parent_child "$json_output" "Pack" "Camera"
    validate_parent_child "$json_output" "Pack" "Clothes"
    validate_parent_child "$json_output" "Pack" "Passport"
    
    # Clean up
    rm -f "$TEST_DIR/create_with_children.sh"
}

@test "create grandchild for one of the children" {
    # Create a test script that adds multiple levels of hierarchy
    cat > "$TEST_DIR/create_grandchild.sh" << 'EOF'
#!/bin/zsh
too add "Project"
too add --to 1 "Frontend"
too add --to 1.1 "Components"
too add --to 1.1.1 "Button.tsx"
too list
EOF
    chmod +x "$TEST_DIR/create_grandchild.sh"
    
    # Run the test script with JSON format
    output=$("$LIVE_TESTS_DIR/run" --format json "$TEST_DIR/create_grandchild.sh")
    
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
    
    # Validate total count
    total_count=$(get_total_count "$json_output")
    [ "$total_count" -eq 4 ]
    
    # Verify the hierarchy
    validate_parent_child "$json_output" "Project" "Frontend"
    validate_parent_child "$json_output" "Frontend" "Components"
    validate_parent_child "$json_output" "Components" "Button.tsx"
    
    # Clean up
    rm -f "$TEST_DIR/create_grandchild.sh"
}