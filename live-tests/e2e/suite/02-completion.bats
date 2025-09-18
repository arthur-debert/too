#!/usr/bin/env bats

# Test suite for completion and clean operations
# Tests: complete child items, clean, complete parent items, --all flag behavior

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

@test "complete a child item" {
    # Create a test script that sets up baseline and completes a child
    cat > "$TEST_DIR/complete_child.sh" << 'EOF'
#!/bin/zsh
# Setup baseline data inline
too add "Groceries"
too add --to 1 "Bread"
too add --to 1 "Milk"
too add --to 1 "Pancakes"
too add --to 1 "Eggs"
too complete 1.4  # Complete Eggs

too add "Pack"
too add --to 2 "Camera"
too add --to 2 "Clothes"
too add --to 2 "Passport"
too add --to 2 "Bag"
too complete 2.4  # Complete Bag

# Complete a child item (Bread)
too complete 1.1

# List with --all to see completed items
too list --all
EOF
    chmod +x "$TEST_DIR/complete_child.sh"
    
    # Run the test script with JSON format
    output=$("$LIVE_TESTS_DIR/run" --format json "$TEST_DIR/complete_child.sh")
    
    # Extract the list --all command JSON output
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
    
    # Verify the Bread item is completed
    bread_todo=$(get_todo_by_text "$json_output" "Bread")
    bread_status=$(echo "$bread_todo" | jq -r '.statuses.completion')
    [ "$bread_status" = "done" ]
    
    # Verify we have 3 completed items total (Eggs, Bag, and now Bread)
    done_count=$(get_done_count "$json_output")
    [ "$done_count" -eq 3 ]
    
    # Clean up
    rm -f "$TEST_DIR/complete_child.sh"
}

@test "completed items visible with --all but not without" {
    # Create a test script that demonstrates visibility of completed items
    cat > "$TEST_DIR/completed_visibility.sh" << 'EOF'
#!/bin/zsh
# Setup baseline data inline
too add "Groceries"
too add --to 1 "Bread"
too add --to 1 "Milk"
too add --to 1 "Pancakes"
too add --to 1 "Eggs"
too complete 1.4  # Complete Eggs

too add "Pack"
too add --to 2 "Camera"
too add --to 2 "Clothes"
too add --to 2 "Passport"
too add --to 2 "Bag"
too complete 2.4  # Complete Bag

# List without --all (should not show completed items)
too list
echo "SEPARATOR"
# List with --all (should show completed items)
too list --all
EOF
    chmod +x "$TEST_DIR/completed_visibility.sh"
    
    # Run the test script with JSON format
    output=$("$LIVE_TESTS_DIR/run" --format json "$TEST_DIR/completed_visibility.sh")
    
    # Split output by SEPARATOR
    list_output=$(echo "$output" | awk '/^{/{p=1} p{print} /SEPARATOR/{exit}' | awk '
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
    
    list_all_output=$(echo "$output" | awk '/SEPARATOR/{p=1;next} p' | awk '
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
    
    # Without --all: should not see Eggs or Bag
    ! todo_exists "$list_output" "Eggs"
    ! todo_exists "$list_output" "Bag"
    
    # With --all: should see Eggs and Bag  
    todo_exists "$list_all_output" "Eggs"
    todo_exists "$list_all_output" "Bag"
    
    # Clean up
    rm -f "$TEST_DIR/completed_visibility.sh"
}

@test "clean removes completed items" {
    # Create a test script that tests clean functionality
    cat > "$TEST_DIR/test_clean.sh" << 'EOF'
#!/bin/zsh
# Setup baseline data inline
too add "Groceries"
too add --to 1 "Bread"
too add --to 1 "Milk"
too add --to 1 "Pancakes"
too add --to 1 "Eggs"
too complete 1.4  # Complete Eggs

too add "Pack"
too add --to 2 "Camera"
too add --to 2 "Clothes"
too add --to 2 "Passport"
too add --to 2 "Bag"
too complete 2.4  # Complete Bag

# Complete one more item to have something to clean
too complete 1.2 # Complete Milk

# Clean completed items
too clean

# List with --all to verify completed items are gone
too list --all
EOF
    chmod +x "$TEST_DIR/test_clean.sh"
    
    # Run the test script with JSON format
    output=$("$LIVE_TESTS_DIR/run" --format json "$TEST_DIR/test_clean.sh")
    
    # Extract the final list --all JSON output
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
    
    # Verify completed items are removed
    ! todo_exists "$json_output" "Eggs"
    ! todo_exists "$json_output" "Bag"
    ! todo_exists "$json_output" "Milk"
    
    # Verify remaining items are still there
    todo_exists "$json_output" "Groceries"
    todo_exists "$json_output" "Bread"
    todo_exists "$json_output" "Pancakes"
    
    # Verify done count is 0 after clean
    done_count=$(get_done_count "$json_output")
    [ "$done_count" -eq 0 ]
    
    # Clean up
    rm -f "$TEST_DIR/test_clean.sh"
}

@test "complete parent preserves children status" {
    # Create a test script that tests parent completion behavior (status only bubbles UP, not DOWN)
    cat > "$TEST_DIR/complete_parent.sh" << EOF
#!/bin/zsh
# Setup baseline data inline since source won't work in the isolated environment
too add "Groceries" --format "\${TOO_FORMAT}"
too add --to 1 "Bread" --format "\${TOO_FORMAT}"
too add --to 1 "Milk" --format "\${TOO_FORMAT}"
too add --to 1 "Pancakes" --format "\${TOO_FORMAT}"
too add --to 1 "Eggs" --format "\${TOO_FORMAT}"
too complete 1.4 --format "\${TOO_FORMAT}"  # Complete Eggs

too add "Pack" --format "\${TOO_FORMAT}"
too add --to 2 "Camera" --format "\${TOO_FORMAT}"
too add --to 2 "Clothes" --format "\${TOO_FORMAT}"
too add --to 2 "Passport" --format "\${TOO_FORMAT}"
too add --to 2 "Bag" --format "\${TOO_FORMAT}"
too complete 2.4 --format "\${TOO_FORMAT}"  # Complete Bag

# Complete parent item (Pack) - should NOT auto-complete children (status only bubbles UP)
too complete 2 --format "\${TOO_FORMAT}"

# List with --all to see all completed items
too list --all --format "\${TOO_FORMAT}"
EOF
    chmod +x "$TEST_DIR/complete_parent.sh"
    
    # Run the test script with JSON format
    output=$("$LIVE_TESTS_DIR/run" --format json "$TEST_DIR/complete_parent.sh")
    
    # Extract the list --all command JSON output
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
    
    # Verify parent is completed
    pack_todo=$(get_todo_by_text "$json_output" "Pack")
    pack_status=$(echo "$pack_todo" | jq -r '.statuses.completion')
    [ "$pack_status" = "done" ]
    
    # Children should preserve their individual statuses (NOT auto-completed)
    camera_todo=$(get_todo_by_text "$json_output" "Camera")
    camera_status=$(echo "$camera_todo" | jq -r '.statuses.completion')
    [ "$camera_status" = "pending" ]
    
    clothes_todo=$(get_todo_by_text "$json_output" "Clothes")
    clothes_status=$(echo "$clothes_todo" | jq -r '.statuses.completion')
    [ "$clothes_status" = "pending" ]
    
    passport_todo=$(get_todo_by_text "$json_output" "Passport")
    passport_status=$(echo "$passport_todo" | jq -r '.statuses.completion')
    [ "$passport_status" = "pending" ]
    
    # Bag was already completed in baseline and should remain completed
    bag_todo=$(get_todo_by_text "$json_output" "Bag")
    bag_status=$(echo "$bag_todo" | jq -r '.statuses.completion')
    [ "$bag_status" = "done" ]
    
    # Done count should be 3 (Eggs, Bag, Pack) - children preserve their status
    done_count=$(get_done_count "$json_output")
    [ "$done_count" -eq 3 ]
    
    # Clean up
    rm -f "$TEST_DIR/complete_parent.sh"
}

@test "reopen completed items" {
    # Create a test script that tests reopening functionality
    cat > "$TEST_DIR/test_reopen.sh" << 'EOF'
#!/bin/zsh
# Simple test case for reopen
too add "Task 1"
too add "Task 2"
too add "Task 3"

# Complete the first task
too complete 1

# List with --all to see the completed task
too list --all

# Reopen the completed task - it should be c1
too reopen c1

# List to verify it's back to pending
too list
EOF
    chmod +x "$TEST_DIR/test_reopen.sh"
    
    # Run the test script with JSON format
    output=$("$LIVE_TESTS_DIR/run" --format json "$TEST_DIR/test_reopen.sh")
    
    # Extract the final list JSON output
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
    
    # Verify Task 1 is back and pending
    task1_todo=$(get_todo_by_text "$json_output" "Task 1")
    task1_status=$(echo "$task1_todo" | jq -r '.statuses.completion')
    [ "$task1_status" = "pending" ]
    
    # All tasks should now be pending (done count = 0)
    done_count=$(get_done_count "$json_output")
    [ "$done_count" -eq 0 ]
    
    # Total count should still be 3
    total_count=$(get_total_count "$json_output")
    [ "$total_count" -eq 3 ]
    
    # Clean up
    rm -f "$TEST_DIR/test_reopen.sh"
}