#!/bin/bash

# Build script for Autosphere MCP Golang Server

echo "🏗️  Building Autosphere MCP Server..."

# Clean previous builds
rm -f autosphere-mcp-server

# Download dependencies
echo "📦 Downloading dependencies..."
go mod tidy

# Build the binary
echo "🔨 Building binary..."
go build -o autosphere-mcp-server ./cmd/server

if [ $? -eq 0 ]; then
    echo "✅ Build successful!"
    echo "🚀 Binary created: ./autosphere-mcp-server"
    echo ""
    echo "Usage:"
    echo "  STDIO transport:      ./autosphere-mcp-server"
    echo "  HTTP transport:       ./autosphere-mcp-server -http localhost:8080"
    echo ""
    echo "Testing:"
    echo "  STDIO with inspector: ./test.sh stdio"
    echo "  HTTP server:          ./test.sh http [port]"
    echo "  HTTP inspector:       ./test.sh inspect [port]"
    echo ""
    echo "Make test script executable:"
    chmod +x test.sh
    echo "  chmod +x test.sh"
else
    echo "❌ Build failed!"
    exit 1
fi
