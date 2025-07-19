# go-template-project

> A batteries-included Go starter template that replicates your Python collaboration template's friction-free developer experience.

Transform your team's Go development from "blank `go mod init` and vibes" to "clone, rename, ship" without bikeshedding file layout, lint rules, or Docker optimization tricks.

## Quick Start

```bash
git clone https://github.com/your-org/go-template-project.git my-new-project
cd my-new-project
go run scripts/init.go    # Interactive project setup
make setup               # Install development tools  
make check               # Verify everything works
go run ./cmd/cli         # Test the CLI
```

## What You Get

**Zero-config development environment:**
- One-command project initialization
- Quality gates that match your CI pipeline
- Pre-commit hooks prevent broken commits
- 80% test coverage enforcement
- Security scanning built in

**Production-ready container deployment:**
- Multi-stage Dockerfile → ~10MB distroless images
- Cross-platform builds for Linux/macOS/Windows
- Docker Compose for local development

**Comprehensive documentation system:**
- Hugo-powered static site generation
- Auto-generated API docs from Go code
- MkDocs-style developer experience
- GitHub Pages deployment ready
- Health checks and graceful shutdown

**Complete CI/CD automation:**
- GitHub Actions for testing, security, and release
- Codecov integration for coverage tracking
- Automated vulnerability scanning
- SARIF integration for GitHub Security tab

## Interactive Initialization

The `scripts/init.go` script handles everything your Python template's `make init` does:

- **Project customization**: Name, module path, description
- **Component selection**: CLI, HTTP server, background worker, docs
- **Git integration**: Repository setup with initial commit
- **Import path updates**: Automatic code generation
- **Pre-commit hooks**: Quality enforcement from day one

```bash
🚀 Go Template Project Initialization
=====================================

Project name [my-new-project]: awesome-service
Go module path [github.com/your-org/awesome-service]: github.com/myorg/awesome-service
Project description: A microservice for awesome things
Author name [John Doe]: Jane Developer
Author email [john@example.com]: jane@myorg.com
License [MIT]: MIT

Components to include:
Include CLI application [Y/n]: y
Include HTTP server [Y/n]: y  
Include background worker [y/N]: n
Include documentation setup [Y/n]: y

✅ Project initialized successfully!
```

## Available Commands

### Development Workflow
```bash
make setup          # Install development tools
make check          # Complete quality gate (fmt + vet + lint + test + security + coverage)  
make ci             # Full CI pipeline locally

# Individual quality checks
make fmt            # Format code with gofumpt
make vet            # Static analysis with go vet
make lint           # Comprehensive linting with golangci-lint
make test           # Tests with race detection and coverage
make security       # Vulnerability and security scanning
make coverage       # Enforce 80% coverage minimum
```

### Testing Categories
```bash
make test-unit      # Fast unit tests
make test-integration  # Component interaction tests
make test-smoke     # Critical path validation
make test-all       # All test categories
```

### Build and Run
```bash
make build          # Build all binaries
make build-all      # Cross-platform builds
make run-cli        # Run CLI application
make run-server     # Run HTTP server
make run-worker     # Run background worker
```

### Container Operations
```bash
make docker-build   # Build optimized Docker images
make docker-run     # Run server in container
make docker-dev     # Start full development environment
```

### Documentation
```bash
make docs-setup     # Install Hugo and gomarkdoc
make docs-serve     # Start local docs server (http://localhost:1313)
make docs-generate  # Generate API docs from Go code
make docs-build     # Build static site for deployment
```

## Project Structure

Following the Standard Go Project Layout with some enhancements:

```
my-project/
├── cmd/                     # One binary per subdirectory
│   ├── cli/                 # Command-line interface
│   ├── server/              # HTTP server
│   └── worker/              # Background worker
├── internal/                # Private application code
│   ├── app/                 # Core business logic
│   ├── config/              # Configuration management
│   └── handlers/            # HTTP request handlers
├── scripts/                 # Development and build scripts
│   └── init.go              # Interactive project initialization
├── .github/workflows/       # CI/CD automation
│   ├── ci.yml               # Main CI pipeline
│   ├── security.yml         # Security scanning
│   └── release.yml          # Automated releases
├── docker/                  # Container configuration
│   ├── docker-compose.yml   # Development environment
│   └── init.sql             # Database initialization
├── docs/                    # Documentation
├── Dockerfile               # Multi-stage container builds
├── Makefile                 # Development workflow automation
└── README.md                # Generated project documentation
```

## Quality Standards

Matches your Python template's quality-first approach:

- **80% test coverage minimum** (configurable)
- **Zero tolerance for linting errors**
- **Pre-commit hooks mirror CI pipeline**
- **Security scanning on every commit**
- **No commits bypass quality gates**

## Container Architecture

Multi-stage Dockerfile produces three optimized images:

```dockerfile
# CLI image (~10MB)
FROM gcr.io/distroless/static-debian12:nonroot AS cli

# Server image (~10MB)  
FROM gcr.io/distroless/static-debian12:nonroot AS server

# Worker image (~10MB)
FROM gcr.io/distroless/static-debian12:nonroot AS worker
```

Benefits:
- **Minimal attack surface**: Distroless base images
- **Small footprint**: ~10MB per image vs ~100MB+ typical Go images
- **Security**: No shell, package managers, or unnecessary binaries
- **Performance**: Fast startup and low memory usage

## CI/CD Pipeline

Three-workflow approach for comprehensive automation:

### 1. CI Workflow (`.github/workflows/ci.yml`)
- **Matrix testing**: Go 1.22 and 1.23
- **Quality gates**: Format, vet, lint, test, security, coverage
- **Build verification**: All binaries compile successfully
- **Container testing**: Images build and run correctly
- **Integration testing**: End-to-end validation

### 2. Security Workflow (`.github/workflows/security.yml`)
- **Vulnerability scanning**: Official `govulncheck` tool
- **Security analysis**: `gosec` static analysis
- **Container scanning**: Trivy vulnerability detection
- **Dependency review**: GitHub dependency scanning
- **SARIF integration**: Results in GitHub Security tab

### 3. Release Workflow (`.github/workflows/release.yml`)
- **Cross-platform builds**: Linux, macOS, Windows (AMD64 + ARM64)
- **GitHub releases**: Automated with checksums
- **Container registry**: Multi-platform images to GitHub Packages
- **Semantic versioning**: Automatic version management

## Getting Started

### 1. Create Your Project
```bash
git clone https://github.com/your-org/go-template-project.git my-new-service
cd my-new-service
go run scripts/init.go
```

### 2. Set Up Development Environment
```bash
make setup           # Install development tools
make check           # Verify quality gates pass
```

### 3. Start Developing
```bash
# Make your changes
vim internal/app/app.go

# Verify quality
make check

# Test locally
make run-cli
make run-server
```

### 4. Deploy
```bash
# Build containers
make docker-build

# Or use pre-built images
docker run ghcr.io/your-org/my-new-service:latest
```

## Configuration

All applications support configuration via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `HOST` | `0.0.0.0` | HTTP server bind address |
| `DEBUG` | `false` | Enable debug logging |
| `DATABASE_URL` | | Database connection string |
| `READ_TIMEOUT` | `15s` | HTTP read timeout |
| `WRITE_TIMEOUT` | `15s` | HTTP write timeout |

## Comparison to Python Template

| Feature | Python Template | Go Template |
|---------|-----------------|-------------|
| **Setup** | `make init` | `go run scripts/init.go` |
| **Quality Gate** | `make check` (ruff + mypy + pytest) | `make check` (gofumpt + vet + golangci-lint + test + security) |
| **Container Size** | ~100MB | ~10MB |
| **Build Time** | 30-60s | 5-15s |
| **Startup Time** | 1-3s | <100ms |
| **Memory Usage** | 50-100MB | 5-20MB |
| **Deployment** | Python runtime required | Single static binary |

## Migration from Python Template

Moving from your Python collaboration template? Here's the mapping:

1. **Project structure**: Similar philosophy, Go-specific layout
2. **Quality gates**: Same rigor, Go-native tools
3. **Pre-commit hooks**: Same enforcement, Go-focused checks  
4. **CI/CD pipeline**: Equivalent GitHub Actions workflows
5. **Container strategy**: Multi-stage builds, much smaller images
6. **Documentation**: Same natural writing style, Go examples

## 🔮 Future Enhancements

**Inspired by [SchwarzIT/go-template](https://github.com/SchwarzIT/go-template) analysis:**

### Phase 1: Container & Release Optimization
- **Container optimization**: Add `go.uber.org/automaxprocs` for automatic GOMAXPROCS configuration
- **GoReleaser integration**: Automated cross-platform releases with GitHub releases
- **Go Report Card**: External code quality validation and badge

### Phase 2: Modern Go Practices  
- **Structured logging**: Migrate from `log` to `log/slog` for better observability
- **CLI distribution tool**: Create `gt new` command for easier project generation (similar to `create-react-app`)
- **Production-ready defaults**: Enhanced signal handling and graceful shutdown patterns

### Phase 3: Developer Experience
- **Simplified project structure**: Reduce complexity while maintaining flexibility
- **Template variants**: Industry-specific templates (microservice, API gateway, worker queue)
- **IDE integration**: VS Code extension for template management
- **Plugin system**: Extensible component architecture

## Contributing

1. Fork and clone the repository
2. Run `make setup` to install development tools
3. Make your changes with tests
4. Run `make check` to verify quality
5. Submit a pull request

## License

MIT - see LICENSE file for details.

---

## Why This Template Exists

Your Python collaboration template succeeds because it eliminates decision fatigue and provides immediate productivity. This Go template brings the same "clone, rename, ship" experience to Go development while leveraging Go's unique strengths:

- **Static compilation**: No runtime dependencies
- **Fast builds**: Sub-15-second CI pipelines  
- **Tiny containers**: 10MB vs 100MB+ Python images
- **Memory efficiency**: 5-20MB vs 50-100MB+ Python
- **Performance**: Microsecond startup times

The result: a Go development experience that feels as smooth as your Python template but ships faster, runs cheaper, and scales better.

*Generated from [go-template-project](https://github.com/your-org/go-template-project) - A batteries-included Go starter template inspired by the python-collab-template philosophy.*