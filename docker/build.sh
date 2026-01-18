#!/bin/sh
set -e

echo "Building NOFX for Linux AMD64..."

# Ensure we're building for the correct platform
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64

echo "Environment variables:"
echo "CGO_ENABLED=$CGO_ENABLED"
echo "GOOS=$GOOS"
echo "GOARCH=$GOARCH"

# Go build with static linking
go env
go build -a -tags netgo -ldflags='-extldflags "-static" -s -w' -o nofx .

echo "Build completed successfully!"