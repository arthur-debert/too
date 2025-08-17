#!/bin/bash
set -e

echo "Testing with real ed editor..."

# Create ed commands file
cat > /tmp/ed-commands << 'EOF'
a
My todo created with ed
This is line two
And line three
.
w
q
EOF

# Use the integration environment with real ed
./scripts/run-integration-env.sh << 'COMMANDS'
echo "=== Testing add with real ed editor ==="
export EDITOR=ed
# Use ed in script mode with commands
tdh add --editor < /tmp/ed-commands
tdh list
exit
COMMANDS

rm -f /tmp/ed-commands