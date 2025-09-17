E2E Test Suite

The e2e test suite provides comprehensive integration testing for too's 
command-line interface using Bats (Bash Automated Testing System).

Structure:
    suite/          - Main test files
    utils/          - JSON parsing and test utilities
    fixtures/       - Test data setup scripts

Running Tests:

    # Run all tests
    ./run-tests.sh
    
    # Run specific test file
    bats suite/01-creation.bats
    
    # Run single test
    bats suite/01-creation.bats -f "create item at top level"
    
    # Run with different output format
    TOO_FORMAT=json ./run-tests.sh

Results:
    - Human-readable output to stdout
    - JUnit XML for CI (in temp dir or $E2E_RESULTS_DIR)
    - TAP format for compatibility

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