---
title: "Getting Started"
linkTitle: "Getting Started"
weight: 1
description: "Complete setup guide for the Go Template Project"
---

# Getting Started

This guide will help you set up and customize the Go Template Project for your needs.

## Prerequisites

Ensure you have the following installed:

- **Go 1.23+** - [Install from golang.org](https://golang.org/downloads/)
- **Git** - For version control
- **Make** - For build automation
- **Docker** (optional) - For containerized development

```bash
# Verify installations
go version    # Should show Go 1.23+
git --version
make --version
docker --version
```

## Quick Setup

### 1. Clone the Template

```bash
git clone https://github.com/your-org/go-template-project.git my-awesome-project
cd my-awesome-project
```

### 2. Initialize Your Project

The interactive initialization script will customize the template:

```bash
go run scripts/init.go
```

This will prompt you for:
- **Project name** - Your application name
- **Module path** - Go module identifier (e.g., github.com/yourorg/project)
- **Description** - Brief project description
- **Author information** - Your name and email
- **License** - Choose from common licenses
- **Applications** - Which example apps to keep (CLI, server, worker)
- **E2E tests** - Whether to include end-to-end tests

### 3. Install Development Tools

```bash
make setup
```

This installs essential Go development tools:
- `golangci-lint` - Comprehensive linting
- `gofumpt` - Stricter formatting than gofmt
- `govulncheck` - Vulnerability scanning
- `gosec` - Security analysis

### 4. Verify Setup

Run the complete quality gate to ensure everything works:

```bash
make check
```

This runs:
- Code formatting (`make fmt`)
- Static analysis (`make vet`)
- Linting (`make lint`)
- Tests with coverage (`make test`)
- Security scans (`make security`)
- Coverage validation (`make coverage`)

## Project Structure

After initialization, your project will have this structure:

```
my-awesome-project/
├── cmd/                    # Application entry points
│   ├── cli/               # Command-line application
│   ├── server/            # HTTP server application
│   └── worker/            # Background worker application
├── internal/              # Private application code
│   ├── cli/              # CLI-specific logic
│   ├── server/           # Server-specific logic
│   ├── worker/           # Worker-specific logic
│   └── shared/           # Shared utilities
├── tests/                 # Test files
│   └── e2e/              # End-to-end tests
├── scripts/              # Build and utility scripts
├── docs/                 # Documentation source
├── docker/               # Docker configurations
├── .github/              # GitHub Actions workflows
├── Makefile              # Build automation
├── Dockerfile            # Multi-stage build
├── go.mod                # Go module definition
└── README.md             # Project documentation
```

## Development Workflow

### Daily Development

1. **Make changes** to your code
2. **Run quality checks** frequently:
   ```bash
   make fmt    # Format code
   make test   # Run tests
   make lint   # Check for issues
   ```
3. **Build and test** your applications:
   ```bash
   make build
   make run-cli      # Test CLI
   make run-server   # Test server (in another terminal)
   make run-worker   # Test worker
   ```

### Quality Gates

The template enforces quality through multiple gates:

#### Pre-commit Hooks
Automatically installed during `make setup`:
- Code formatting validation
- Basic linting checks
- Test execution

#### Comprehensive Checks
Run the full quality gate before commits:
```bash
make check
```

#### Continuous Integration
GitHub Actions automatically run on pull requests:
- Multi-platform testing (Linux, macOS, Windows)
- Security vulnerability scanning
- Code coverage reporting
- Docker image building

## Building Applications

### Local Development

```bash
# Build all applications
make build

# Run specific applications
make run-cli
make run-server
make run-worker

# Cross-platform builds
make build-all
```

### Docker Development

```bash
# Build Docker image
make docker-build

# Run in container
make docker-run

# Development environment with hot reload
make docker-dev
```

## Testing Strategy

The template implements a progressive testing approach:

### Test Categories

```bash
# Unit tests (fast, isolated)
make test-unit

# Integration tests (component interactions)
make test-integration

# Smoke tests (critical path validation)
make test-smoke

# End-to-end tests (complete user journeys)
make test-e2e

# All tests
make test-all
```

### Coverage Requirements

- **Minimum coverage**: Configurable in Makefile (template defaults to 0%)
- **Coverage report**: Generated in `coverage.out`
- **HTML report**: `go tool cover -html=coverage.out`

## Customizing the Template

### Adding New Applications

1. Create new directory in `cmd/`:
   ```bash
   mkdir cmd/myapp
   ```

2. Add main.go with your application logic

3. Update Makefile with build targets:
   ```makefile
   build:
       CGO_ENABLED=0 go build -o bin/myapp ./cmd/myapp

   run-myapp:
       go run ./cmd/myapp
   ```

### Modifying Quality Standards

Edit the Makefile to adjust quality requirements:

```makefile
# Change coverage threshold
COVERAGE_MIN ?= 90

# Modify linting timeout
GOLANGCI_FLAGS ?= --timeout=10m

# Customize test flags
GOTEST_FLAGS ?= -race -covermode=atomic -timeout=30s
```

### Updating Dependencies

```bash
# Add new dependencies
go get github.com/example/package

# Update all dependencies
go get -u all
make tidy

# Verify no vulnerabilities
make security
```

## Next Steps

Now that your project is set up:

1. **Review the [Architecture Guide](architecture/)** to understand design patterns
2. **Explore [Examples](../examples/)** for common implementation patterns
3. **Set up [CI/CD](deployment/)** for your repository
4. **Customize [Configuration](configuration/)** for your environment

## Troubleshooting

### Common Issues

**Go modules not working:**
```bash
go mod tidy
go mod download
```

**Linter failures:**
```bash
make fmt      # Fix formatting
make lint     # See specific issues
```

**Test failures:**
```bash
make test-unit     # Run unit tests only
go test -v ./...   # Verbose test output
```

**Docker build issues:**
```bash
make clean         # Clean build artifacts
make docker-build  # Rebuild image
```

### Getting Help

- Check [troubleshooting guide](troubleshooting/) for detailed solutions
- Review [API documentation](../api/) for package details
- Open an issue on [GitHub](https://github.com/your-org/go-template-project/issues)
