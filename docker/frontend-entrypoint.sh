#!/bin/bash
set -e

echo "Installing missing rollup binaries..."
npm install --no-save @rollup/rollup-linux-x64-gnu || echo "Failed to install @rollup/rollup-linux-x64-gnu, continuing..."

echo "Starting Vite development server..."
exec "$@"