#!/bin/bash
set -e

echo "Testing editor functionality with ed..."

# Create a script that will simulate ed editor commands
cat > /tmp/ed-test-script << 'EOF'
#!/bin/bash
# This script simulates ed editor interaction
tmpfile="$1"

# If file has content, we're in edit mode
if [ -s "$tmpfile" ]; then
    # Simulate editing existing content
    cat > "$tmpfile" << 'CONTENT'
Edited todo with multiple lines
This is the second line
And a third line for good measure
CONTENT
else
    # Simulate adding new content
    cat > "$tmpfile" << 'CONTENT'


  My new todo item
  With some extra lines...
  
  And blank lines in between!
  

CONTENT
fi
EOF

chmod +x /tmp/ed-test-script

# Use the integration environment with our fake ed
EDITOR=/tmp/ed-test-script ./scripts/run-integration-env.sh << 'COMMANDS'
echo "=== Testing add with editor ==="
tdh add --editor
tdh list

echo -e "\n=== Testing add with editor and initial content ==="
tdh add --editor "Initial content to edit"
tdh list

echo -e "\n=== Testing edit with editor ==="
tdh edit 1 --editor
tdh list

echo -e "\n=== Testing add to parent with editor ==="
tdh add 1 --editor
tdh list

echo -e "\n=== Testing with --format markdown ==="
tdh list --format markdown

exit
COMMANDS

# Clean up
rm -f /tmp/ed-test-script