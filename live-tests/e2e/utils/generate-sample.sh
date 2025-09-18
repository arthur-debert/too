#!/bin/bash

# Sample todo list generator for e2e tests
# Creates common todo list structures for testing

set -euo pipefail

# Generate a simple flat todo list
# Usage: generate_flat_list <count>
generate_flat_list() {
    local count=${1:-3}
    
    for i in $(seq 1 "$count"); do
        echo "too add \"Todo Item $i\" --format \"\${TOO_FORMAT}\""
    done
}

# Generate a hierarchical todo list
# Usage: generate_hierarchical_list
generate_hierarchical_list() {
    cat << 'EOF'
too add "Groceries"
too add --to 1 "Milk"
too add --to 1 "Bread"
too add --to 1 "Eggs"
too add "Pack for Trip"
too add --to 2 "Clothes"
too add --to 2 "Camera Gear"
too add --to 2 "Passport"
EOF
}

# Generate a list with mixed statuses (some completed)
# Usage: generate_mixed_status_list
generate_mixed_status_list() {
    cat << 'EOF'
too add "Task 1"
too add "Task 2"
too add "Task 3"
too add "Task 4"
too complete 1
too complete 3
EOF
}

# Generate a test script for basic operations
# Usage: generate_basic_operations_script <output_file>
generate_basic_operations_script() {
    local output_file="$1"
    
    cat > "$output_file" << 'EOF'
#!/bin/zsh

# Create some todos
too add "Buy groceries"
too add "Call dentist"
too add "Plan weekend trip"

# Create hierarchical structure
too add --to 1 "Apples"
too add --to 1 "Bread"
too add --to 3 "Book hotel"
too add --to 3 "Check weather"

# Complete some tasks
too complete 1.1  # Complete "Apples"
too complete 2    # Complete "Call dentist"

# List final state
too list
EOF
    
    chmod +x "$output_file"
}

# Generate a test script for edge cases
# Usage: generate_edge_cases_script <output_file>
generate_edge_cases_script() {
    local output_file="$1"
    
    cat > "$output_file" << 'EOF'
#!/bin/zsh

# Test empty list
too list

# Add todo with special characters
too add "Todo with \"quotes\" and 'apostrophes'"
too add "Todo with unicode: 🚀 🎯 ✅"

# Test very long text
too add "This is a very long todo item that has a lot of text to test how the system handles longer descriptions and whether everything works correctly with extended content"

# Test hierarchy limits
too add "Level 1"
too add --to 1 "Level 2"
too add --to 1.1 "Level 3"

# List final state
too list
EOF
    
    chmod +x "$output_file"
}

# If script is run directly, show usage
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    echo "Sample todo list generator for e2e tests"
    echo ""
    echo "Available functions:"
    echo "  generate_flat_list <count>            - Generate simple flat list"
    echo "  generate_hierarchical_list            - Generate hierarchical structure"  
    echo "  generate_mixed_status_list            - Generate list with completed items"
    echo "  generate_basic_operations_script <file> - Generate comprehensive test script"
    echo "  generate_edge_cases_script <file>     - Generate edge case test script"
    echo ""
    echo "Usage example:"
    echo "  source generate-sample.sh"
    echo "  generate_flat_list 5 > test_script.sh"
    echo "  chmod +x test_script.sh"
fi