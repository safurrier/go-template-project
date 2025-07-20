# Architecture Guide

This document explains the architectural decisions and patterns used in this Go template project.

## Design Philosophy

This template follows the **MODEST** principles for sustainable software architecture:

- **Modularity**: Building components that are swappable and reusable
- **Orthogonality**: Components are independent, changes are localized
- **Dependency Injection**: External dependencies are passed in explicitly
- **Explicitness**: Intent is clear in code, avoiding magic behavior
- **Single Responsibility**: Each component has one reason to change
- **Testability**: Code is designed to be easily testable

## Project Structure

### Standard Go Project Layout

We follow the widely-adopted [Standard Go Project Layout](https://github.com/golang-standards/project-layout) with enhancements:

```
project/
├── cmd/                    # Application entry points
├── internal/               # Private application code
├── pkg/                    # Public library code (when needed)
├── scripts/                # Build and development scripts
├── .github/workflows/      # CI/CD automation
└── docker/                 # Container configuration
```

### Directory Purposes

#### `/cmd`
- **One binary per subdirectory**: Each folder becomes a separate executable
- **Minimal main.go**: Delegates to internal packages for testability
- **Examples**: `cli/`, `server/`, `worker/`

#### `/internal`
- **Private application code**: Cannot be imported by external projects
- **Business logic**: Core functionality lives here
- **Organized by domain**: `app/`, `config/`, `handlers/`

#### `/pkg` (Optional)
- **Public APIs**: Code intended for external use
- **Stable interfaces**: Versioned and backward-compatible
- **Libraries**: Reusable components for other projects

## Component Architecture

### Application Layer (`internal/app`)

The core application logic is centralized in the `app` package:

```go
type App struct {
    Name    string
    Version string
    Debug   bool
}

func (a *App) Run() error {
    // Business logic here
    return nil
}
```

**Benefits:**
- Testable without main()
- Consistent across CLI/server/worker
- Easy to extend with new functionality

### Configuration (`internal/config`)

Environment-based configuration with sensible defaults:

```go
type Config struct {
    Port         int
    Host         string
    Debug        bool
    ReadTimeout  time.Duration
    WriteTimeout time.Duration
}

func Load() (*Config, error) {
    // Load from environment with defaults
}
```

**Patterns:**
- Fail fast on invalid configuration
- Environment variables override defaults
- Type-safe configuration structs

### HTTP Handlers (`internal/handlers`)

Clean, testable HTTP handlers with proper separation:

```go
func HealthCheck(version string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Handler logic
    }
}
```

**Benefits:**
- Easy to test with httptest
- Dependency injection for testability
- Consistent error handling

## Testing Strategy

### Progressive Testing Approach

Following the user's established testing conventions:

1. **End-to-End Tests**: Validate complete user workflows
2. **Smoke Tests**: Verify critical paths work
3. **Integration Tests**: Test component interactions
4. **Unit Tests**: Test individual functions and methods

### E2E Testing Architecture

The template includes comprehensive E2E testing for real-world validation:

#### Test Categories
- **CLI Tests**: Application launch, version flags, help system, error handling
- **Server Tests**: HTTP server lifecycle, health endpoints, graceful shutdown
- **Worker Tests**: Background processing, task intervals, signal handling
- **Init Tests**: Interactive setup, project customization, git integration

#### Key Testing Principles
- **Real processes**: Execute actual binaries, not mocks
- **Signal handling**: Use `SIGTERM` for reliable process termination
- **Configurable timing**: Environment variables for test intervals (`WORKER_TASK_INTERVAL`)
- **Environment isolation**: Custom `TMPDIR` for NAS/restricted filesystems
- **Timeout protection**: Prevent hanging tests with disciplined cleanup

### Test Organization

```
project/
├── internal/app/app_test.go        # Unit tests
├── internal/config/config_test.go   # Unit tests
├── internal/handlers/health_test.go # Unit tests
└── tests/
    ├── integration/                 # Integration tests
    ├── smoke/                      # Smoke tests
    └── e2e/                        # End-to-end tests
```

### Testing Principles

- **Test behavior, not implementation**
- **Avoid mocks in favor of test doubles**
- **One assertion per test**
- **Descriptive test names**

## Container Architecture

### Multi-Stage Dockerfile

Three optimized images from one Dockerfile:

```dockerfile
# Build stage
FROM golang:1.23-alpine AS builder

# CLI Runtime
FROM gcr.io/distroless/static-debian12:nonroot AS cli

# Server Runtime
FROM gcr.io/distroless/static-debian12:nonroot AS server

# Worker Runtime
FROM gcr.io/distroless/static-debian12:nonroot AS worker
```

### Benefits

- **Security**: Distroless images have minimal attack surface
- **Size**: ~10MB images vs ~100MB+ typical Go containers
- **Performance**: Fast startup, low memory usage
- **Flexibility**: Different images for different use cases

## Quality Architecture

### Quality Gates

Every commit must pass:

1. **Formatting**: `gofumpt` (stricter than `gofmt`)
2. **Static Analysis**: `go vet` for correctness
3. **Linting**: `golangci-lint` with comprehensive rules
4. **Testing**: Unit tests with race detection
5. **Security**: `govulncheck` and `gosec` scanning
6. **Coverage**: Configurable threshold (template defaults to 0% for flexibility)

### Pre-commit Hooks

Local quality enforcement mirrors CI pipeline:

```yaml
repos:
  - repo: local
    hooks:
      - id: go-fmt
        name: Format Go code
        entry: make
        language: system
        args: [fmt]
        types: [go]
        pass_filenames: false

      - id: golangci-lint
        name: golangci-lint
        entry: make
        language: system
        args: [lint]
        types: [go]
        pass_filenames: false
```

### CI/CD Pipeline

Three workflows for comprehensive automation:

1. **CI**: Testing, building, quality gates
2. **Security**: Vulnerability scanning, SARIF integration
3. **Release**: Cross-platform builds, container publishing

## Dependency Management

### Tool Dependencies in go.mod

Go 1.21+ supports tool dependencies directly:

```go
require (
    github.com/golangci/golangci-lint v1.56.0
    mvdan.cc/gofumpt v0.6.0
    golang.org/x/vuln/cmd/govulncheck v1.0.0
    github.com/securecodewarrior/gosec/v2/cmd/gosec v2.18.2
)
```

### Benefits

- **Version pinning**: Consistent tool versions across team
- **No tools.go hacks**: Native Go module support
- **Automatic installation**: `go install` handles dependencies

## Security Architecture

### Built-in Security

- **Vulnerability scanning**: Official `govulncheck` integration
- **Static analysis**: `gosec` security linting
- **Container scanning**: Trivy vulnerability detection
- **Dependency review**: GitHub security monitoring

### Secure Defaults

- **Distroless containers**: No shell or package managers
- **Non-root user**: All containers run as unprivileged user
- **Minimal dependencies**: Only essential packages included
- **Regular updates**: Automated dependency and base image updates

## Performance Characteristics

### Build Performance

- **Incremental builds**: Go module caching
- **Parallel compilation**: Multi-core utilization
- **Fast CI**: 5-15 second build times

### Runtime Performance

- **Static compilation**: No runtime dependencies
- **Low memory usage**: 5-20MB typical footprint
- **Fast startup**: <100ms cold start
- **Efficient networking**: Go's excellent HTTP/2 support

## Extensibility Patterns

### Adding New Commands

1. Create directory in `cmd/`: `cmd/newcommand/`
2. Add `main.go` that delegates to `internal/app`
3. Update Makefile with new build targets
4. Add component to `scripts/init.go` selection

### Adding New Internal Packages

1. Create package in `internal/`: `internal/newpackage/`
2. Follow MODEST principles for design
3. Add comprehensive tests
4. Update documentation

### Adding Dependencies

1. Use `go get` for runtime dependencies
2. Add to `require` section in go.mod
3. For tools, add to tool dependencies section
4. Update CI workflows if needed

## Migration Considerations

### From Python Template

Key differences when migrating from Python collaboration template:

- **Build artifacts**: Binaries instead of Python packages
- **Dependencies**: go.mod instead of requirements.txt/pyproject.toml
- **Quality tools**: golangci-lint instead of ruff/mypy
- **Containers**: Static binaries instead of Python runtime

### Gradual Adoption

- Start with one component (CLI or server)
- Migrate shared libraries to internal packages
- Update CI/CD pipeline incrementally
- Maintain same quality standards throughout

## Troubleshooting

### Common Issues

1. **Import cycles**: Reorganize package dependencies
2. **Test failures**: Check for race conditions with `-race`
3. **Container size**: Verify multi-stage build stages
4. **CI performance**: Review caching strategies

### E2E Test Issues

Based on debugging experience, common E2E test problems and solutions:

#### Test Hanging or Resource Exhaustion
- **Symptom**: Tests hang indefinitely, high CPU/memory usage
- **Cause**: Git operations with incompatible commit messages, recursive pre-commit hooks
- **Solution**: Use conventional commit format, verify pre-commit hook calls tool not itself

#### Process Termination Issues
- **Symptom**: E2E tests fail to terminate server/worker processes
- **Cause**: Using `os.Interrupt` instead of `syscall.SIGTERM`
- **Solution**: Switch to `SIGTERM` with proper exit code handling (143/130)

#### Execution Permission Errors
- **Symptom**: Tests fail on NAS or restricted filesystems
- **Cause**: Default temp directory lacks execution permissions
- **Solution**: Use custom `TMPDIR` environment variable

#### Worker Test Timing
- **Symptom**: Worker tests fail due to timing mismatches
- **Cause**: Default 10-second task interval too slow for tests
- **Solution**: Use `WORKER_TASK_INTERVAL=2s` environment variable

#### Pre-commit Hook Recursion
- **Symptom**: Infinite loop during commit operations
- **Cause**: Global hook calls itself instead of pre-commit tool
- **Solution**: Verify hook calls `pre-commit run --hook-stage pre-commit`

### Debug Tools

- **Delve debugger**: Advanced debugging capabilities
- **pprof profiling**: Performance analysis
- **Build flags**: `-ldflags` for optimization
- **Verbose output**: `go build -v` for build details

This architecture provides a solid foundation for Go projects while maintaining the same developer experience quality as your Python collaboration template.
