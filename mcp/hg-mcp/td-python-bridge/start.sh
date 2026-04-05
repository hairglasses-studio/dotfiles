#!/bin/bash
# TouchDesigner Python Bridge - Startup Script

set -e

cd "$(dirname "$0")"

echo "TouchDesigner Python Bridge - Starting..."
echo ""

# Check if virtual environment exists
if [ ! -d "venv" ]; then
    echo "Creating Python virtual environment..."
    python3 -m venv venv
fi

# Activate virtual environment
echo "Activating virtual environment..."
source venv/bin/activate

# Install dependencies
echo "Installing dependencies..."
pip install -q --upgrade pip
pip install -q -r requirements.txt

echo ""
echo "Starting bridge server..."
echo ""

# Run the bridge
python bridge.py
