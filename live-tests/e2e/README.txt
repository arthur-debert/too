E2E Test Suite

The e2e test suite provides comprehensive integration testing for too's 
command-line interface using Bats (Bash Automated Testing System).

Structure:
    suite/          - Main test files
    utils/          - JSON parsing and test utilities
    fixtures/       - Test data setup scripts

Running Tests:

    # Run all tests (default: human-friendly output)
    ./run-tests.sh
    
    # Run with JUnit XML output (for CI)
    ./run-tests.sh --output junit
    
    # Run with TAP output
    ./run-tests.sh --output tap
    
    # Run specific test file
    bats suite/01-creation.bats
    
    # Run single test
    bats suite/01-creation.bats -f "create item at top level"
    
    # Run with different too output format
    TOO_FORMAT=json ./run-tests.sh

Results:
    - nice: Human-readable output with colors and timing (default)
    - junit: JUnit XML for CI systems (saved to temp dir or $E2E_RESULTS_DIR)
    - tap: TAP format for compatibility (to stdout)

Best Practices:
    - Use baseline fixture for consistent test data
    - Parse JSON output for reliable assertions
    - Test JSON format unless testing display
    - Keep tests independent and idempotent
    - Use parse-json.sh utilities for validation
    - One test file per functional area

Test Files:
    01-creation.bats     - Todo creation operations
    02-completion.bats   - Complete, reopen, clean
    03-text-content.bats - Special text handling
    04-edit-move.bats    - Edit and move operations

Utilities:
    parse-json.sh - Extract and validate JSON data
    generate-sample.sh - Create test data structures