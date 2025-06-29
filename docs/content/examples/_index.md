---
title: "Examples"
linkTitle: "Examples" 
weight: 30
description: "Code examples and implementation patterns"
---

# Examples

This section provides practical examples and implementation patterns for common Go development scenarios using the template.

## Application Examples

The template includes three complete example applications that demonstrate different architectural patterns:

### [CLI Application](cli/)
Command-line interface with flags, subcommands, and proper error handling.

```bash
./bin/cli --version
./bin/cli --help
./bin/cli process --input data.txt --output results.json
```

**Key Patterns:**
- Flag parsing and validation
- Subcommand architecture
- Error handling and user feedback
- Configuration from environment/files

### [HTTP Server](server/)
REST API server with middleware, routing, and graceful shutdown.

```bash
curl http://localhost:8080/health
curl http://localhost:8080/api/info
```

**Key Patterns:**
- HTTP routing and middleware
- Request/response handling
- Graceful shutdown with context
- Health checks and monitoring

### [Background Worker](worker/)
Long-running process with signal handling and configuration.

```bash
DEBUG=true ./bin/worker
```

**Key Patterns:**
- Signal handling (SIGINT, SIGTERM)
- Background task processing
- Logging and monitoring
- Configuration management

## Code Patterns

### [Error Handling](error-handling/)
Comprehensive error handling strategies for Go applications.

### [Configuration](configuration/)
Environment-based configuration with validation and defaults.

### [Testing Patterns](testing/)
Progressive testing approach with practical examples.

### [Logging](logging/)
Structured logging with different levels and outputs.

### [Docker Integration](docker/)
Container deployment and development workflows.

## Quick Reference

### Common Tasks

```bash
# Development workflow
make fmt test lint     # Quality checks
make build run-cli     # Build and test

# Testing different scenarios  
make test-unit         # Fast unit tests
make test-e2e         # Full integration

# Docker development
make docker-build      # Build container
make docker-run        # Run in container
```

### Project Customization

```bash
# Initialize new project
make init

# Add new application
mkdir cmd/newapp
# Create main.go
# Update Makefile

# Add dependencies
go get github.com/example/pkg
make tidy
```

### CI/CD Integration

```bash
# Local CI simulation
make ci               # Run full CI pipeline

# Release preparation
git tag v1.0.0
git push origin v1.0.0  # Triggers release workflow
```

## Best Practices

### Code Organization
- Keep `main.go` minimal - business logic in `internal/`
- Use dependency injection for testability
- Separate configuration from implementation
- Follow single responsibility principle

### Error Handling
- Return errors, don't panic in library code
- Wrap errors with context using `fmt.Errorf`
- Handle errors at appropriate boundaries
- Provide actionable error messages

### Testing
- Start with smoke tests for critical paths
- Use table-driven tests for multiple scenarios
- Mock external dependencies, not internal logic
- Test behavior, not implementation

### Performance
- Profile before optimizing
- Use `context.Context` for cancellation
- Prefer simple solutions over complex optimizations
- Monitor memory allocations in hot paths

## Next Steps

- Explore the [API Reference](../api/) for detailed package documentation
- Review [Architecture Guide](../docs/architecture/) for design patterns
- Check [Deployment Guide](../docs/deployment/) for production setup