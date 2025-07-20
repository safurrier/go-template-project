# Development Setup Guide

This document explains how to set up the development environment for the Go template project.

## Prerequisites

- Go 1.23 or later
- Git
- Make
- Docker (optional, for container builds)

## Initial Setup

1. **Clone the repository**:
   ```bash
   git clone https://github.com/your-org/go-template-project.git
   cd go-template-project
   ```

2. **Install development tools**:
   ```bash
   make setup
   ```
   This installs:
   - golangci-lint for linting
   - gofumpt for formatting
   - gomarkdoc for documentation
   - Hugo for docs generation

3. **Verify setup**:
   ```bash
   make check
   ```

## Pre-commit Hooks

### Installation

The template includes pre-commit hooks that run the same quality checks as CI:

1. **Install pre-commit** (if not already installed):
   ```bash
   pip install pre-commit
   # or
   pipx install pre-commit
   ```

2. **Install hooks**:
   ```bash
   pre-commit install
   ```

3. **Test hooks**:
   ```bash
   pre-commit run --all-files
   ```

### Configuration

The pre-commit configuration (`.pre-commit-config.yaml`) includes:
- Go code formatting with gofumpt
- Static analysis with `go vet`
- Linting with golangci-lint (using `make lint`)
- Go module tidying
- YAML/JSON validation
- End-of-file and whitespace fixes
- Commit message validation

### Troubleshooting Pre-commit Hooks

**If hooks fail or hang:**
```bash
# Check installation
pre-commit --version

# Reinstall hooks
pre-commit uninstall
pre-commit install

# Run manually
pre-commit run --all-files
```

**For NAS/restricted environments:**
The hooks are configured to work with `CGO_ENABLED=0` and handle execution permission issues automatically.

## Development Workflow

### Daily Development
```bash
# 1. Make changes to code
# 2. Run quality checks
make check

# 3. Run tests
make test

# 4. Commit (pre-commit hooks run automatically)
git commit -m "feat: add new feature"
```

### Testing

**Run all tests:**
```bash
make test-all
```

**Run specific test categories:**
```bash
make test-unit        # Unit tests
make test-integration # Integration tests
make test-smoke       # Smoke tests
make test-e2e         # E2E tests
```

**For E2E tests on NAS/restricted filesystems:**
```bash
mkdir -p ~/tmp
TMPDIR=$HOME/tmp make test-e2e
```

### Building

**Build all binaries:**
```bash
make build
```

**Cross-platform builds:**
```bash
make build-all
```

**Docker containers:**
```bash
make docker-build
```

### Running Applications

**CLI application:**
```bash
make run-cli
# or
go run ./cmd/cli
```

**HTTP server:**
```bash
make run-server
# or
go run ./cmd/server
```

**Background worker:**
```bash
make run-worker
# or
DEBUG=true WORKER_TASK_INTERVAL=5s go run ./cmd/worker
```

## Environment Configuration

All applications support configuration via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `DEBUG` | `false` | Enable debug logging |
| `PORT` | `8080` | HTTP server port |
| `HOST` | `0.0.0.0` | HTTP server bind address |
| `WORKER_TASK_INTERVAL` | `10s` | Worker task processing interval |

## Documentation

**Start documentation server:**
```bash
make docs-serve
```

**Generate API documentation:**
```bash
make docs-generate
```

**Build static documentation:**
```bash
make docs-build
```

## CI/CD Integration

The project includes GitHub Actions workflows:
- **CI**: Quality gates, testing, and build validation
- **Documentation**: Automatic documentation deployment
- All workflows match local development tools

## Common Issues

### Go Version Mismatch
Ensure Dockerfile and go.mod use the same Go version:
```bash
grep "go 1" go.mod
grep "golang:" Dockerfile
```

### Test Execution Permissions
For NAS or restricted environments:
```bash
mkdir -p ~/tmp
export TMPDIR=$HOME/tmp
```

### Worker Test Timing
Worker tests use configurable intervals:
```bash
WORKER_TASK_INTERVAL=2s go test -tags=e2e ./tests/e2e/worker_e2e_test.go
```

### Pre-commit Hook Recursion
If you encounter infinite loops, check that the global pre-commit hook calls the tool, not itself:
```bash
cat ~/.git/hooks/pre-commit
# Should call: pre-commit run --hook-stage pre-commit
# Not: ./.git/hooks/pre-commit "$@"
```

## Best Practices

1. **Always run `make check` before committing**
2. **Use descriptive commit messages** following conventional commits
3. **Add tests for new functionality**
4. **Update documentation** when adding features
5. **Use environment variables** for configuration instead of hardcoding values
6. **Follow the MODEST principles** in code design

## Getting Help

- Check the main [README.md](../README.md) for overview
- See [ARCHITECTURE.md](./ARCHITECTURE.md) for design patterns and testing guidance
- File issues on GitHub for bugs or feature requests
