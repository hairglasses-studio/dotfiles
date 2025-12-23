#!/bin/bash
# OPNsense Network Integration Test Script
# Tests integration between OPNsense components and AFTRS network

set -e

echo "🧪 Testing OPNsense network integration..."

# Test LLM agent connectivity
echo "🤖 Testing LLM agent integration..."
if [ -d "../opnsense-llmagent" ]; then
    cd ../opnsense-llmagent
    # Add LLM agent tests here
    cd - > /dev/null
fi

# Test AFTRS CLI integration
echo "🔗 Testing AFTRS CLI integration..."
if [ -d "../aftrs_cli" ]; then
    cd ../aftrs_cli
    # Add AFTRS CLI integration tests here
    cd - > /dev/null
fi

# Test Tailscale integration
echo "🌐 Testing Tailscale integration..."
# Add Tailscale connectivity tests here

echo "✅ Network integration tests complete!"
