#!/bin/sh
set -e

echo "Starting initialization script..."

# Check if air is available
if ! command -v air >/dev/null 2>&1; then
  echo "Error: air command not found"
  exit 1
fi

# Copy .air.conf to a temporary location where we have write permissions
CONFIG_PATH="/tmp/air.conf.copy"
cp /tmp/.air.conf $CONFIG_PATH
chmod 644 $CONFIG_PATH

echo "Air is available, starting..."

# Run air with the copied configuration file
exec air -c $CONFIG_PATH "$@"