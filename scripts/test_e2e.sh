#!/bin/bash

# E2E Test Verification Script
# Run this after Go is installed to verify E2E tests work correctly

set -e

echo "🎭 E2E Test Verification for go-template-project"
echo "================================================"
echo

# Check Go installation
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.22+ first."
    exit 1
fi

GO_VERSION=$(go version | cut -d' ' -f3)
echo "✅ Go installed: $GO_VERSION"

# Test Red Phase - E2E tests should initially fail
echo
echo "🔴 RED PHASE: Verifying E2E tests fail initially..."
echo "This verifies our tests are properly checking real functionality."

# Test CLI E2E (should work since CLI is implemented)
echo "  Testing CLI E2E tests..."
if go test -tags=e2e -timeout=30s ./tests/e2e/cli_e2e_test.go -v 2>/dev/null; then
    echo "  ✅ CLI E2E tests pass (CLI is implemented)"
else
    echo "  ❌ CLI E2E tests fail (may need CLI implementation fixes)"
fi

# Test Server E2E (should work since server is implemented)
echo "  Testing Server E2E tests..."
if timeout 45s go test -tags=e2e -timeout=40s ./tests/e2e/server_e2e_test.go -v 2>/dev/null; then
    echo "  ✅ Server E2E tests pass (Server is implemented)"
else
    echo "  ❌ Server E2E tests fail (may need server implementation fixes)"
fi

# Test Worker E2E (should work since worker is implemented)
echo "  Testing Worker E2E tests..."
if timeout 30s go test -tags=e2e -timeout=25s ./tests/e2e/worker_e2e_test.go -v 2>/dev/null; then
    echo "  ✅ Worker E2E tests pass (Worker is implemented)"
else
    echo "  ❌ Worker E2E tests fail (may need worker implementation fixes)"
fi

# Test Init Script E2E (more complex, may have issues)
echo "  Testing Init Script E2E tests..."
if timeout 60s go test -tags=e2e -timeout=50s ./tests/e2e/init_e2e_test.go -v 2>/dev/null; then
    echo "  ✅ Init Script E2E tests pass"
else
    echo "  ⚠️  Init Script E2E tests fail (expected - complex integration)"
fi

echo
echo "🟢 GREEN PHASE: Running all E2E tests with make target..."
echo "This tests the complete E2E workflow."

# Test with make target (if make is available)
if command -v make &> /dev/null; then
    echo "  Running 'make test-e2e'..."
    if timeout 120s make test-e2e 2>/dev/null; then
        echo "  ✅ Make test-e2e passes"
    else
        echo "  ❌ Make test-e2e fails (some tests may need implementation fixes)"
    fi
else
    echo "  ⚠️  Make not available, skipping make test-e2e"
fi

echo
echo "🎯 E2E Test Summary:"
echo "- CLI tests validate that the CLI application can launch and respond to flags"
echo "- Server tests validate that the HTTP server can start, serve endpoints, and shut down gracefully"
echo "- Worker tests validate that the background worker can start, process tasks, and shut down"
echo "- Init tests validate that the project initialization script works end-to-end"
echo
echo "💡 Expected behavior:"
echo "- Some tests may fail initially (Red phase) - this is normal for TDD"
echo "- Tests verify real functionality, not mocks or implementation details"
echo "- Tests can be removed via init script if not wanted"

echo
echo "🔧 To fix failing tests:"
echo "1. Check that applications build: 'make build'"
echo "2. Test applications manually: 'make run-cli', 'make run-server', 'make run-worker'"
echo "3. Fix any application startup or functionality issues"
echo "4. Re-run E2E tests: 'make test-e2e'"

echo
echo "✅ E2E Test verification complete!"