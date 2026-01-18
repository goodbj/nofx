#!/bin/sh
set -e

echo "Starting NOFX application..."
echo "Current architecture: $(uname -m)"
echo "Current OS: $(uname -s)"

# Check the binary file
echo "Checking binary file..."
ls -la /app/nofx
file /app/nofx 2>/dev/null || echo "file command not available"

# Make sure the binary is executable
chmod +x /app/nofx

# Verify it's executable
if [ -x /app/nofx ]; then
    echo "Binary is executable, starting application..."
    exec /app/nofx
else
    echo "Binary is not executable!"
    exit 1
fi