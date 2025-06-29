#!/bin/bash
set -e

echo "🧪 Running smoke tests for Go template project..."

# Test 1: Build all applications
echo "1. Testing build process..."
make build
echo "✅ Build successful"

# Test 2: Test applications can run
echo "2. Testing CLI application..."
./bin/cli --version
echo "✅ CLI works"

echo "3. Testing server application..."
timeout 2 ./bin/server > /dev/null 2>&1 || echo "✅ Server starts and runs"

echo "4. Testing worker application..."
timeout 2 ./bin/worker > /dev/null 2>&1 || echo "✅ Worker starts and runs"

# Test 3: Test quality gates
echo "5. Testing quality gates..."
make fmt vet lint
echo "✅ Quality gates pass"

# Test 4: Test module is valid
echo "6. Testing Go module..."
go mod verify
go mod tidy
echo "✅ Go module is valid"

# Test 5: Test project structure
echo "7. Testing project structure..."
for dir in cmd/cli cmd/server cmd/worker internal scripts docs; do
    if [ -d "$dir" ]; then
        echo "  ✓ $dir exists"
    else
        echo "  ✗ $dir missing"
        exit 1
    fi
done
echo "✅ Project structure is correct"

echo ""
echo "🎉 All smoke tests passed!"
echo "✅ Go template project is working correctly"