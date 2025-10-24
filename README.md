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
go run ./cmd/cli --help  # Discover the live telemetry dashboard options
```

## What You Get

**Zero-config development environment:**
- One-command project initialization
- Quality gates that match your CI pipeline
- Pre-commit hooks prevent broken commits
- Comprehensive testing with E2E validation
- Docker builds with Go 1.23 support

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
make check          # Complete quality gate (fmt + vet + lint + test)
make ci             # Full CI pipeline locally

# Individual quality checks
make fmt            # Format code with gofumpt
make vet            # Static analysis with go vet
make lint           # Comprehensive linting with golangci-lint
make test           # Tests with race detection and coverage
make coverage       # Generate test coverage reports
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

### Live F1 Telemetry Dashboard

The Bubbletea-powered CLI streams timing data from the [OpenF1 API](https://api.openf1.org) and presents it with live updates.

```bash
# Follow the latest session and focus on driver VER
go run ./cmd/cli --driver VER

# Inspect a historic session
go run ./cmd/cli --session 9896 --refresh 10s

# Override the telemetry API endpoint
go run ./cmd/cli --api https://api.openf1.org/v1 --driver 44
```

Use the arrow keys to move between drivers, `q` to quit, and the `--refresh` flag to control the polling interval. When live data is unavailable the dashboard stays open and continues to poll automatically.

#### undercut-f1 parity features

The dashboard now mirrors the flagship views from [undercut-f1](https://github.com/JustAman62/undercut-f1) so you can keep tabs on a session without leaving the terminal:

- **Timing Tower** – tabular view with tyre compound and age, last/best lap, sector splits, cumulative interval, gap to the leader, and relative delta to the currently selected driver.
- **Race Control** – chronological feed of steward messages with lap references, flag state, and affected drivers.
- **Tyre Strategy** – stint breakdown per driver highlighting compound changes and lap ranges.
- **Driver Tracker** – gap-based track strip that highlights each driver, making it easy to visualise pit windows without GPS data.
- **Lap History** – rolling table of the latest laps for the focused driver including deltas to the previous and best laps plus sector performance.

The tracker view approximates car placement using timing gaps because OpenF1 does not expose live GPS coordinates, and team radio playback remains out of scope for now.

### Phased Feature Plan

To progressively mirror the experience provided by [undercut-f1](https://github.com/JustAman62/undercut-f1), the dashboard evolves through the following phases:

1. **Phase 1 – Multi-view timing workspace**
   - Introduce a tabbed layout with a persistent header/footer.
   - Preserve the timing tower as the primary view while preparing the model for additional panes.
   - Add keyboard navigation for switching between views without disrupting driver selection.
2. **Phase 2 – Race control stream**
   - Ingest race control messages from the OpenF1 feed.
   - Surface categorized messages in a dedicated view with clear timestamping and driver attribution.
   - Refresh the feed on the same cadence as the timing tower to keep control updates current.
3. **Phase 3 – Tyre strategy overview**
   - Fetch tyre stint data for every driver and group it alongside the timing tower roster.
   - Display compound, lap range, and tyre age details so offset strategies are obvious at a glance.
   - Keep strategy information synchronized with timing updates so it stays relevant throughout the session.
4. **Phase 4 – Driver tracker strip**
   - Recreate undercut-f1’s tracker view by projecting gaps onto a track strip and highlighting the active driver.
   - Surface gap-to-leader and gap-to-selected metrics beside each driver entry.
5. **Phase 5 – Lap history timeline**
   - Provide a lap-by-lap table with delta-to-previous and delta-to-best comparisons alongside sector splits.
   - Automatically refresh the history view when the focused driver completes a lap.

All five phases are implemented in this iteration so the CLI delivers timing, race control, strategy, tracker, and lap history insights in a single cohesive interface.

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

- **Comprehensive testing** with unit, integration, and E2E tests
- **Zero tolerance for linting errors**
- **Pre-commit hooks mirror CI pipeline**
- **Consistent formatting with gofumpt**
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
- **Go 1.23 support**: Latest Go version with enhanced tooling
- **Quality gates**: Format, vet, lint, test, coverage
- **Build verification**: All binaries compile successfully
- **Container testing**: Images build and run correctly
- **Integration testing**: End-to-end validation including init script

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
| `WORKER_TASK_INTERVAL` | `10s` | Worker task processing interval |

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

## Troubleshooting

### Pre-commit Hooks
If pre-commit hooks fail or seem to hang:
```bash
# Check pre-commit installation
pre-commit --version

# Reinstall hooks
pre-commit uninstall
pre-commit install

# Test hooks manually
pre-commit run --all-files
```

### E2E Tests
If E2E tests hang or fail:
```bash
# Run with custom temp directory (for NAS/restricted environments)
mkdir -p ~/tmp
TMPDIR=$HOME/tmp make test-e2e

# Run specific test
TMPDIR=$HOME/tmp go test -tags=e2e -run TestInitScript ./tests/e2e/
```

### Docker Build Issues
If Docker builds fail with Go version errors:
```bash
# Verify Go version in Dockerfile matches go.mod
grep "go 1" go.mod
grep "golang:" Dockerfile
```

### Worker Configuration
The worker supports configurable task intervals for testing:
```bash
# Fast testing interval
WORKER_TASK_INTERVAL=2s DEBUG=true go run ./cmd/worker
```

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
