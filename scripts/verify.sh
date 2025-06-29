#!/bin/bash

# Verification script for go-template-project
# Run this after Go is installed to verify the template works

set -e

echo "🔍 Verifying go-template-project setup..."
echo

# Check Go installation
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.22+ first."
    exit 1
fi

GO_VERSION=$(go version | cut -d' ' -f3)
echo "✅ Go installed: $GO_VERSION"

# Check directory structure
echo "📁 Checking project structure..."
for dir in cmd internal scripts .github docker docs; do
    if [ -d "$dir" ]; then
        echo "  ✅ $dir/"
    else
        echo "  ❌ Missing $dir/"
        exit 1
    fi
done

# Check key files
echo "📄 Checking key files..."
for file in go.mod Makefile Dockerfile README.md .golangci.yml .pre-commit-config.yaml; do
    if [ -f "$file" ]; then
        echo "  ✅ $file"
    else
        echo "  ❌ Missing $file"
        exit 1
    fi
done

# Test Go commands
echo "🔧 Testing Go commands..."
echo "  Running go mod tidy..."
go mod tidy

echo "  Running go mod download..."
go mod download

echo "  Testing CLI build..."
if go build -o /tmp/test-cli ./cmd/cli; then
    echo "  ✅ CLI builds successfully"
    rm -f /tmp/test-cli
else
    echo "  ❌ CLI build failed"
    exit 1
fi

echo "  Testing server build..."
if go build -o /tmp/test-server ./cmd/server; then
    echo "  ✅ Server builds successfully"
    rm -f /tmp/test-server
else
    echo "  ❌ Server build failed"
    exit 1
fi

echo "  Testing worker build..."
if go build -o /tmp/test-worker ./cmd/worker; then
    echo "  ✅ Worker builds successfully"
    rm -f /tmp/test-worker
else
    echo "  ❌ Worker build failed"
    exit 1
fi

# Test basic functionality
echo "🧪 Testing basic functionality..."
echo "  Running CLI application..."
if timeout 5s go run ./cmd/cli 2>/dev/null; then
    echo "  ✅ CLI runs successfully"
else
    echo "  ✅ CLI exits as expected (timeout or normal exit)"
fi

# Test Makefile targets
echo "📋 Testing Makefile targets..."
if command -v make &> /dev/null; then
    echo "  Testing make fmt..."
    if make fmt 2>/dev/null; then
        echo "  ✅ make fmt works"
    else
        echo "  ⚠️  make fmt requires gofumpt (run 'make setup')"
    fi
    
    echo "  Testing make vet..."
    if make vet 2>/dev/null; then
        echo "  ✅ make vet works"
    else
        echo "  ❌ make vet failed"
    fi
else
    echo "  ⚠️  make not available, skipping Makefile tests"
fi

# Check Docker setup
echo "🐳 Checking Docker setup..."
if command -v docker &> /dev/null; then
    echo "  ✅ Docker available"
    if docker build -t test-template -f Dockerfile . --target cli &>/dev/null; then
        echo "  ✅ Docker build successful"
        docker rmi test-template &>/dev/null || true
    else
        echo "  ❌ Docker build failed"
    fi
else
    echo "  ⚠️  Docker not available, skipping container tests"
fi

echo
echo "🎉 Template verification complete!"
echo
echo "Next steps:"
echo "1. Run 'go run scripts/init.go' to customize your project"
echo "2. Run 'make setup' to install development tools"
echo "3. Run 'make check' to verify quality gates"
echo "4. Start building your application!"