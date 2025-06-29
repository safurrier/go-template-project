---
title: "Go Template Project"
linkTitle: "Go Template"
description: "A batteries-included Go starter template for rapid development"
---

# Go Template Project

Welcome to the Go Template Project - a comprehensive, batteries-included starter template for Go applications.

## 🚀 Quick Start

```bash
# Clone and initialize your project
git clone https://github.com/your-org/go-template-project.git my-project
cd my-project
make init

# Start developing
make setup
make check
make build
```

## ✨ Features

This template provides everything you need to start building production-ready Go applications:

- **🏗️ Standard Project Layout** - Follows Go community best practices
- **🔧 Complete Toolchain** - golangci-lint, gofumpt, govulncheck, gosec
- **🧪 Testing Framework** - Unit, integration, smoke, and E2E tests
- **📦 Docker Support** - Multi-stage builds with distroless images (~10MB)
- **🚀 CI/CD Ready** - GitHub Actions for testing, security, and releases
- **📚 Documentation** - Auto-generated API docs with Hugo
- **⚡ Quality Gates** - Pre-commit hooks and comprehensive checks

## 🏛️ Architecture

The template includes three example applications demonstrating different patterns:

- **CLI Application** (`cmd/cli`) - Command-line interface with flags and subcommands
- **HTTP Server** (`cmd/server`) - REST API with graceful shutdown and health checks
- **Background Worker** (`cmd/worker`) - Long-running process with signal handling

## 🎯 Getting Started

1. **Initialize Your Project**
   ```bash
   make init
   ```
   Interactive setup will customize the template for your project.

2. **Install Development Tools**
   ```bash
   make setup
   ```
   Installs all necessary Go tools and dependencies.

3. **Run Quality Checks**
   ```bash
   make check
   ```
   Runs the complete quality gate: formatting, linting, testing, security.

4. **Build and Run**
   ```bash
   make build
   make run-cli     # or run-server, run-worker
   ```

## 📖 Documentation

- [Getting Started Guide](docs/getting-started/) - Detailed setup instructions
- [Architecture Overview](docs/architecture/) - Design patterns and structure
- [API Reference](api/) - Auto-generated Go documentation
- [Examples](examples/) - Code examples and tutorials

## 🛠️ Development

This template follows the **MODEST** principles for maintainable Go code:

- **Modularity** - Reusable, swappable components
- **Orthogonality** - Independent components with localized changes
- **Dependency Injection** - External dependencies passed explicitly
- **Explicitness** - Clear intent without magic
- **Single Responsibility** - Each component has one reason to change
- **Testability** - Designed for easy testing

## 🧪 Testing Strategy

Progressive testing approach for sustainable development:

```
End-to-End Tests (Essential user journeys)
    ↓
Smoke Tests (System health validation)
    ↓
Integration Tests (Component interactions)
    ↓
Unit Tests (Individual component behavior)
```

## 📊 Quality Metrics

- **Code Coverage**: Minimum 80% with `make coverage`
- **Linting**: golangci-lint with comprehensive rule set
- **Security**: govulncheck and gosec scanning
- **Dependencies**: Automated vulnerability checking

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Run `make check` to ensure quality
4. Submit a pull request

## 📝 License

This template is available under the MIT License. See [LICENSE](LICENSE) for details.

---

**Ready to build something amazing?** Start with `make init` and let the template guide you!