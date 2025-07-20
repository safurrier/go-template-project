# End-to-End Testing Guide

This document explains the E2E testing strategy for the Go template project.

## Overview

The E2E tests validate complete user journeys through the entire application stack. They test from the user's perspective with real processes, network calls, and system integration.

## Testing Philosophy

Following the progressive testing approach:
1. **E2E Tests** (Start here) - Complete user workflows
2. **Smoke Tests** - Critical path validation
3. **Integration Tests** - Component interactions
4. **Unit Tests** - Individual function testing

## E2E Test Categories

### 1. CLI Application Tests (`cli_e2e_test.go`)

Tests the complete CLI user experience:

- **Application Launch**: Verifies CLI can start without errors
- **Version Flag**: Tests `--version` responds correctly
- **Help System**: Tests `-h` and `--help` provide usage info
- **Error Handling**: Tests invalid flags show helpful errors

**Example User Journey:**
```bash
# User downloads and tries the CLI
./bin/cli --version     # Should show version info
./bin/cli --help        # Should show usage help
./bin/cli --invalid     # Should show helpful error
```

### 2. Server Application Tests (`server_e2e_test.go`)

Tests the complete HTTP server experience:

- **Application Launch**: Verifies server starts and accepts connections
- **Health Endpoints**: Tests `/health` and `/ready` endpoints
- **API Functionality**: Tests `/api/info` endpoint
- **Graceful Shutdown**: Verifies server shuts down properly

**Example User Journey:**
```bash
# User starts the server
./bin/server &

# User checks server health
curl http://localhost:8080/health    # Should return 200 OK
curl http://localhost:8080/api/info  # Should return app info

# User stops server gracefully
kill -INT $SERVER_PID               # Should shut down cleanly
```

### 3. Worker Application Tests (`worker_e2e_test.go`)

Tests the complete background worker experience:

- **Application Launch**: Verifies worker starts and processes tasks
- **Task Processing**: Tests worker performs background work
- **Configuration**: Tests environment variables affect behavior
- **Graceful Shutdown**: Verifies worker stops cleanly

**Example User Journey:**
```bash
# User starts background worker
DEBUG=true ./bin/worker &

# Worker processes tasks in background
# User observes worker activity in logs

# User stops worker gracefully
kill -INT $WORKER_PID               # Should finish current task and exit
```

### 4. Initialization Tests (`init_e2e_test.go`)

Tests the complete project setup experience:

- **Interactive Setup**: Verifies init script completes successfully
- **Input Validation**: Tests script validates user input
- **File Generation**: Tests correct files are created/removed
- **Project Customization**: Tests module paths and names update

**Example User Journey:**
```bash
# User clones template and initializes project
git clone go-template-project my-project
cd my-project
go run scripts/init.go

# Interactive setup customizes project
# User gets working project ready for development
make setup && make check
```

## Test Implementation Principles

### 1. Test Behavior, Not Implementation

❌ **Bad**: Check specific log messages
```go
if !strings.Contains(output, "Processing task #42") {
    t.Fatal("Expected specific log message")
}
```

✅ **Good**: Check that behavior occurred
```go
if !containsWorkerActivity(output) {
    t.Fatal("Worker doesn't show signs of activity")
}
```

### 2. Use Real Processes, Not Mocks

❌ **Bad**: Mock the CLI execution
```go
mockCLI := &MockCLI{}
mockCLI.On("Run").Return(nil)
```

✅ **Good**: Execute real CLI
```go
cmd := exec.Command("go", "run", "./cmd/cli")
err := cmd.Run()
```

### 3. Test Complete User Workflows

❌ **Bad**: Test individual functions
```go
func TestHealthCheckFunction(t *testing.T) {
    handler := HealthCheck("1.0.0")
    // Test handler in isolation
}
```

✅ **Good**: Test end-to-end workflow
```go
func TestServerHealthEndpoint(t *testing.T) {
    // Start real server
    // Make HTTP request
    // Verify response
}
```

### 4. Independent Test Execution

Each test:
- Sets up its own environment
- Uses unique ports/directories to avoid conflicts
- Cleans up after itself
- Can run in parallel with other tests

## Running E2E Tests

### Local Development

```bash
# Run all E2E tests
make test-e2e

# Run specific test file
go test -tags=e2e ./tests/e2e/cli_e2e_test.go -v

# Run with timeout for safety
go test -tags=e2e -timeout=60s ./tests/e2e/...

# Skip long-running tests
go test -tags=e2e -short ./tests/e2e/...
```

### CI/CD Integration

E2E tests run automatically in GitHub Actions:

```yaml
- name: Run E2E tests
  run: make test-e2e
```

Tests include timeouts and parallel execution to minimize CI time.

## Test Configuration

### Build Tags

E2E tests use build tags to separate them from unit tests:

```go
// +build e2e

package e2e
```

### Test Timeouts

Different timeout strategies:
- **Individual tests**: 30-60 seconds for complex workflows
- **Make target**: 60 seconds for complete test suite
- **CI timeout**: Configurable based on runner performance

### Environment Variables

Tests respect environment configuration:
- `DEBUG=true`: Enables verbose output for debugging
- `PORT=8081`: Uses non-default ports to avoid conflicts
- `SHORT=true`: Skips long-running tests via `testing.Short()`
- `TMPDIR=$HOME/tmp`: Uses custom temp directory for NAS/restricted filesystems
- `CGO_ENABLED=0`: Disables CGO for cross-platform compatibility
- `WORKER_TASK_INTERVAL=2s`: Faster worker task intervals for testing

## Removing E2E Tests

E2E tests can be removed during project initialization:

```bash
go run scripts/init.go
# ...
Include E2E tests [y/N]: n  # Choose 'n' to remove E2E tests
```

This removes:
- `tests/` directory entirely
- E2E test targets from Makefile
- E2E test steps from CI workflows

## Best Practices

### 1. Test Critical User Paths First

Focus on workflows users will actually perform:
- Starting and stopping applications
- Accessing primary endpoints
- Basic configuration and setup

### 2. Make Tests Resilient

Use flexible assertions that don't break with minor changes:
- Check for presence of key information, not exact strings
- Use timeouts for operations that might be slow
- Retry mechanisms for network operations

### 3. Provide Clear Error Messages

When tests fail, make it easy to understand why:
```go
if !waitForServer(t, serverURL, 10*time.Second) {
    t.Fatalf("Server at %s did not start within 10 seconds", serverURL)
}
```

### 4. Keep Tests Fast

Optimize for quick feedback:
- Use shorter timeouts for local development
- Run tests in parallel where possible
- Skip expensive tests in short mode

## Troubleshooting

### Common Issues

1. **Port Conflicts**: Tests use different ports (8081, 8082, 8083)
2. **Timeouts**: Increase timeout values for slower systems
3. **Process Cleanup**: Tests kill processes in defer blocks
4. **CI Flakiness**: Tests include retry logic and realistic timeouts
5. **Execution Permissions**: Use custom TMPDIR on NAS/restricted filesystems
6. **Signal Handling**: Tests use SIGTERM for reliable process termination
7. **Worker Timing**: Configurable task intervals for faster test execution

### Debugging Failed Tests

```bash
# Run specific test with verbose output
go test -tags=e2e ./tests/e2e/server_e2e_test.go -v -run TestServerHealthEndpoint

# For NAS/restricted filesystems - use custom TMPDIR
mkdir -p ~/tmp
TMPDIR=$HOME/tmp go test -tags=e2e ./tests/e2e/init_e2e_test.go -v

# Check if applications build correctly
make build

# Test applications manually
make run-cli
make run-server

# Test worker with debug output and fast interval
DEBUG=true WORKER_TASK_INTERVAL=2s make run-worker

# Run E2E tests with environment variables
CGO_ENABLED=0 TMPDIR=$HOME/tmp make test-e2e
```

### Expected Test Behavior

- **CLI tests**: Should pass if CLI builds and runs
- **Server tests**: Should pass if server starts and responds to HTTP
- **Worker tests**: Should pass if worker starts and processes tasks
- **Init tests**: May need adjustment based on project structure changes

## Integration with Template Features

### Progressive Testing Philosophy

E2E tests implement the progressive testing approach:
1. Start with E2E tests for critical user workflows
2. Add smoke tests for fast validation
3. Add integration tests for component interactions
4. Add unit tests for detailed function testing

### Removal via Init Script

Users can choose not to include E2E tests:
- Removes complexity for simple projects
- Reduces CI time for teams that prefer different testing strategies
- Maintains focus on core functionality

### CI/CD Integration

E2E tests integrate with the quality gate:
- Run after unit and integration tests
- Provide final validation before deployment
- Include timeout and retry logic for reliability

This E2E testing approach provides confidence that the complete application stack works correctly from the user's perspective while maintaining the flexibility to remove them when not needed.
