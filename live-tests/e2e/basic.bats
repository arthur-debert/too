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