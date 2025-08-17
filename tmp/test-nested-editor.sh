#!/bin/bash
set -e

echo "Testing nested todo with editor..."

# Create a script that will simulate editor
cat > /tmp/test-editor << 'EOF'
#!/bin/bash
tmpfile="$1"
cat > "$tmpfile" << 'CONTENT'
This is a nested todo
created with editor
CONTENT
EOF

chmod +x /tmp/test-editor

# Use the integration environment
EDITOR=/tmp/test-editor ./scripts/run-integration-env.sh << 'COMMANDS'
echo "=== Creating parent todo ==="
tdh add "Parent todo"
tdh list

echo -e "\n=== Adding nested todo with editor ==="
tdh add 1 --editor
tdh list

exit
COMMANDS

# Clean up
rm -f /tmp/test-editor