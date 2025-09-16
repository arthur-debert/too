#!/usr/bin/env bats

# Basic e2e test suite for too
# This uses the live-tests infrastructure to run isolated tests with JSON output

setup() {
    # Get the directory containing this test file
    TEST_DIR="$( cd "$( dirname "$BATS_TEST_FILENAME" )" && pwd )"
    LIVE_TESTS_DIR="$(dirname "$TEST_DIR")"
    PROJECT_ROOT="$(dirname "$LIVE_TESTS_DIR")"
    
    # Export paths for use in tests
    export TEST_DIR
    export LIVE_TESTS_DIR 
    export PROJECT_ROOT
}

@test "hello world stub test" {
    echo "Hello from Bats e2e test suite!"
    echo "PROJECT_ROOT: $PROJECT_ROOT"
    echo "LIVE_TESTS_DIR: $LIVE_TESTS_DIR"
    echo "TEST_DIR: $TEST_DIR"
    
    # Verify the live-tests/run script exists
    [ -f "$LIVE_TESTS_DIR/run" ]
    [ -x "$LIVE_TESTS_DIR/run" ]
    
    echo "✅ Basic setup verified"
}

@test "can run too command with JSON format" {
    # Test that we can actually run too commands through live-tests infrastructure
    
    # Create a simple test script
    cat > "$TEST_DIR/simple_test.sh" << 'EOF'
#!/bin/zsh
too add "Test Item" --format "${TOO_FORMAT}"
EOF
    chmod +x "$TEST_DIR/simple_test.sh"
    
    # Run the test script with JSON format
    output=$("$LIVE_TESTS_DIR/run" --format json "$TEST_DIR/simple_test.sh")
    
    # Extract JSON from the output by finding the block that starts with { and ends with }
    # We'll extract everything between the first standalone { and the last standalone }
    json_output=$(echo "$output" | sed -n '/^{$/,/^}$/p')
    
    # Verify we got JSON output
    echo "$json_output" | jq . >/dev/null  # This will fail if not valid JSON
    
    # Verify the output contains expected JSON fields
    echo "$json_output" | jq -e '.Command' >/dev/null
    echo "$json_output" | jq -e '.AllTodos' >/dev/null
    echo "$json_output" | jq -e '.TotalCount' >/dev/null
    
    # Verify the todo was created with correct text
    todo_text=$(echo "$json_output" | jq -r '.AllTodos[0].text')
    [ "$todo_text" = "Test Item" ]
    
    # Clean up
    rm -f "$TEST_DIR/simple_test.sh"
    
    echo "✅ JSON format test passed"
}