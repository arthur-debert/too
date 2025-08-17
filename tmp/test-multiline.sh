#!/bin/bash
set -e

echo "Testing multi-line todo formatting..."

# Use the integration environment
./scripts/run-integration-env.sh << 'EOF'
tdh add "Hello, item one"
tdh add "Hello, Item two, multiline
continues here"
tdh add "Hello, item three"
tdh list
tdh list --format markdown
exit
EOF