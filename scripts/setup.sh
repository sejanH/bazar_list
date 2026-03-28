#!/bin/bash

# Setup script for Bazar List project

set -e  # Exit on error

echo "🛒 Setting up Bazar List..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21 or higher."
    echo "   Visit: https://golang.org/dl/"
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo "✓ Found Go version: $GO_VERSION"

# Download dependencies
echo "📦 Downloading dependencies..."
go mod download

# Build the project
echo "🔨 Building the project..."
make build

# Create data directory if it doesn't exist
if [ ! -d "data" ]; then
    mkdir -p data
    echo "✓ Created data directory"
fi

# Run a quick test
echo "🧪 Running a quick test..."
./build/bazarlist add "Test Item" --category other
if [ $? -eq 0 ]; then
    echo "✓ Application is working!"
    ./build/bazarlist remove 1 > /dev/null 2>&1
else
    echo "❌ Application test failed"
    exit 1
fi

echo ""
echo "✅ Setup complete!"
echo ""
echo "Quick start commands:"
echo "  ./build/bazarlist add \"Milk\" --category dairy"
echo "  ./build/bazarlist list"
echo "  ./build/bazarlist help"
echo ""
echo "For more information, see README.md and QUICKSTART.md"
