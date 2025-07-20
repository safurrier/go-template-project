package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"
)

// ProjectConfig holds the configuration for project initialization.
type ProjectConfig struct {
	ProjectName    string
	ModulePath     string
	Description    string
	Author         string
	Email          string
	License        string
	EnableCLI      bool
	EnableServer   bool
	EnableWorker   bool
	EnableDocs     bool
	EnableE2ETests bool
	GitRemote      string
}

// TemplateData holds data for template rendering.
type TemplateData struct {
	ProjectConfig
	Year string
}

const (
	defaultLicense = "MIT"
	defaultAuthor  = "Your Name"
	defaultEmail   = "your.email@example.com"

	// Regex patterns for validation
	projectNamePattern = `^[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9]$`
	modulePathPattern  = `^[a-zA-Z0-9][a-zA-Z0-9-_.]*[a-zA-Z0-9]/` +
		`[a-zA-Z0-9][a-zA-Z0-9-_.]*[a-zA-Z0-9]/[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9]$`
)

func main() {
	fmt.Println("🚀 Go Template Project Initialization")
	fmt.Println("=====================================")
	fmt.Println()

	config, err := gatherProjectInfo()
	if err != nil {
		log.Fatalf("Failed to gather project info: %v", err)
	}

	if err := initializeProject(config); err != nil {
		log.Fatalf("Failed to initialize project: %v", err)
	}

	fmt.Println("\n✅ Project initialized successfully!")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Review the generated files")
	fmt.Println("  2. Run 'make setup' to install development tools")
	fmt.Println("  3. Run 'make check' to verify everything works")
	if config.EnableDocs {
		fmt.Println("  4. Update documentation in docs/ to match your project")
		fmt.Println("  5. Start coding!")
	} else {
		fmt.Println("  4. Start coding!")
	}
}

func gatherProjectInfo() (*ProjectConfig, error) {
	reader := bufio.NewReader(os.Stdin)
	config := &ProjectConfig{}

	// Get current directory name as default project name
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}
	defaultProjectName := filepath.Base(cwd)

	// Project name
	config.ProjectName = promptWithDefault(reader, "Project name", defaultProjectName)
	if !isValidProjectName(config.ProjectName) {
		return nil, fmt.Errorf("invalid project name: must contain only letters, numbers, and hyphens")
	}

	// Module path
	defaultModulePath := fmt.Sprintf("github.com/your-org/%s", config.ProjectName)
	config.ModulePath = promptWithDefault(reader, "Go module path", defaultModulePath)
	if !isValidModulePath(config.ModulePath) {
		return nil, fmt.Errorf("invalid module path format")
	}

	// Description
	config.Description = promptWithDefault(reader, "Project description",
		"A Go application built from go-template-project")

	// Try to get git config for defaults
	gitAuthor := getGitConfig("user.name", defaultAuthor)
	gitEmail := getGitConfig("user.email", defaultEmail)

	config.Author = promptWithDefault(reader, "Author name", gitAuthor)
	config.Email = promptWithDefault(reader, "Author email", gitEmail)
	config.License = promptWithDefault(reader, "License", defaultLicense)

	// Components to enable
	fmt.Println("\nComponents to include:")
	config.EnableCLI = promptBool(reader, "Include CLI application", true)
	config.EnableServer = promptBool(reader, "Include HTTP server", true)
	config.EnableWorker = promptBool(reader, "Include background worker", false)
	config.EnableDocs = promptBool(reader, "Include documentation setup", true)
	config.EnableE2ETests = promptBool(reader, "Include E2E tests", false)

	// Git remote (optional)
	config.GitRemote = prompt(reader, "Git remote URL (optional)")

	// Confirmation
	fmt.Println("\n📋 Configuration Summary:")
	fmt.Printf("  Project Name: %s\n", config.ProjectName)
	fmt.Printf("  Module Path:  %s\n", config.ModulePath)
	fmt.Printf("  Description:  %s\n", config.Description)
	fmt.Printf("  Author:       %s <%s>\n", config.Author, config.Email)
	fmt.Printf("  License:      %s\n", config.License)
	fmt.Printf("  Components:   CLI=%t Server=%t Worker=%t Docs=%t E2E=%t\n",
		config.EnableCLI, config.EnableServer, config.EnableWorker, config.EnableDocs, config.EnableE2ETests)

	if !promptBool(reader, "\nProceed with initialization?", false) {
		fmt.Println("❌ Initialization cancelled")
		os.Exit(0)
	}

	return config, nil
}

func initializeProject(config *ProjectConfig) error {
	// Update go.mod
	if err := updateGoMod(config); err != nil {
		return fmt.Errorf("failed to update go.mod: %w", err)
	}

	// Update import paths in all Go files
	if err := updateImportPaths(config); err != nil {
		return fmt.Errorf("failed to update import paths: %w", err)
	}

	// Remove unwanted components
	if err := removeUnwantedComponents(config); err != nil {
		return fmt.Errorf("failed to remove unwanted components: %w", err)
	}

	// Clean up template artifacts
	if err := cleanupTemplateArtifacts(config); err != nil {
		return fmt.Errorf("failed to clean up template artifacts: %w", err)
	}

	// Generate README
	if err := generateReadme(config); err != nil {
		return fmt.Errorf("failed to generate README: %w", err)
	}

	// Initialize git repository (skip in test environments to prevent hanging)
	if os.Getenv("SKIP_GIT_INIT") == "" {
		if err := initializeGit(config); err != nil {
			fmt.Printf("⚠️  Failed to initialize git: %v\n", err)
			fmt.Println("   Continuing without git initialization...")
		}
	} else {
		fmt.Println("ℹ️  Skipping git initialization (test environment)")
	}

	// Install pre-commit hooks
	if err := setupPreCommitHooks(); err != nil {
		fmt.Printf("⚠️  Failed to setup pre-commit hooks: %v\n", err)
		fmt.Println("   You can set them up later with: pre-commit install")
	}

	// Final cleanup: Remove the init script itself
	fmt.Println("🗑️  Removing initialization script...")
	if err := os.Remove("scripts/init.go"); err != nil {
		fmt.Printf("⚠️  Failed to remove init script: %v\n", err)
		fmt.Println("   You can remove it manually: rm scripts/init.go")
	}

	// Remove scripts directory if it's now empty
	if err := removeEmptyDirectory("scripts"); err != nil {
		// Non-critical, just log
		fmt.Printf("ℹ️  Could not remove scripts directory: %v\n", err)
	}

	return nil
}

func updateGoMod(config *ProjectConfig) error {
	goModContent := fmt.Sprintf(`module %s

go 1.23

require (
	// Runtime dependencies will be added as needed
)
`, config.ModulePath)

	return os.WriteFile("go.mod", []byte(goModContent), 0o644)
}

func updateImportPaths(config *ProjectConfig) error {
	oldPath := "github.com/your-org/go-template-project"
	newPath := config.ModulePath

	return filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-Go files
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Read file
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Replace import paths
		newContent := strings.ReplaceAll(string(content), oldPath, newPath)

		// Write back if changed
		if newContent != string(content) {
			return os.WriteFile(path, []byte(newContent), info.Mode())
		}

		return nil
	})
}

func removeUnwantedComponents(config *ProjectConfig) error {
	// Remove CLI if not wanted
	if !config.EnableCLI {
		if err := os.RemoveAll("cmd/cli"); err != nil {
			return err
		}
	}

	// Remove server if not wanted
	if !config.EnableServer {
		if err := os.RemoveAll("cmd/server"); err != nil {
			return err
		}
		if err := os.RemoveAll("internal/handlers"); err != nil {
			return err
		}
	}

	// Remove worker if not wanted
	if !config.EnableWorker {
		if err := os.RemoveAll("cmd/worker"); err != nil {
			return err
		}
	}

	// Remove docs setup if not wanted
	if !config.EnableDocs {
		if err := os.RemoveAll("docs"); err != nil {
			return err
		}
	}

	// Remove E2E tests if not wanted
	if !config.EnableE2ETests {
		if err := os.RemoveAll("tests"); err != nil {
			return err
		}
	}

	return nil
}

func cleanupTemplateArtifacts(config *ProjectConfig) error {
	fmt.Println("🧹 Cleaning up template artifacts...")

	// Always remove template-specific files
	templateFiles := []string{
		"tests/e2e/init_e2e_test.go", // Tests the init script itself
	}

	for _, file := range templateFiles {
		if err := removeFileIfExists(file); err != nil {
			return fmt.Errorf("failed to remove %s: %w", file, err)
		}
	}

	// Remove component-specific E2E tests based on selection
	if config.EnableE2ETests {
		if !config.EnableCLI {
			if err := removeFileIfExists("tests/e2e/cli_e2e_test.go"); err != nil {
				return err
			}
		}
		if !config.EnableServer {
			if err := removeFileIfExists("tests/e2e/server_e2e_test.go"); err != nil {
				return err
			}
		}
		if !config.EnableWorker {
			if err := removeFileIfExists("tests/e2e/worker_e2e_test.go"); err != nil {
				return err
			}
		}
	}

	// Remove template references from documentation
	if config.EnableDocs {
		if err := cleanupDocumentationReferences(config); err != nil {
			return fmt.Errorf("failed to cleanup documentation: %w", err)
		}
	}

	// Final step: Schedule init script for removal (will remove itself at the end)
	// We can't remove it now since we're running from it
	fmt.Println("   ✅ Scheduled init script for removal")

	return nil
}

func removeFileIfExists(filepath string) error {
	if _, err := os.Stat(filepath); err == nil {
		fmt.Printf("   🗑️  Removing %s\n", filepath)
		return os.Remove(filepath)
	} else if !os.IsNotExist(err) {
		return err
	}
	return nil
}

func removeEmptyDirectory(dirpath string) error {
	// Check if directory exists
	if _, err := os.Stat(dirpath); os.IsNotExist(err) {
		return nil
	}

	// Try to remove directory (will only succeed if empty)
	return os.Remove(dirpath)
}

func cleanupDocumentationReferences(config *ProjectConfig) error {
	// Update Hugo documentation files to remove template references
	docFiles := map[string]func(*ProjectConfig) string{
		"docs/content/_index.md":               updateIndexMarkdown,
		"docs/content/docs/getting-started.md": updateGettingStartedMarkdown,
	}

	for file, updateFunc := range docFiles {
		if err := updateDocumentationFile(file, updateFunc, config); err != nil {
			return fmt.Errorf("failed to update %s: %w", file, err)
		}
	}

	return nil
}

func updateDocumentationFile(filepath string, updateFunc func(*ProjectConfig) string, config *ProjectConfig) error {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to update
	}

	newContent := updateFunc(config)
	return os.WriteFile(filepath, []byte(newContent), 0o644)
}

func updateIndexMarkdown(config *ProjectConfig) string {
	return fmt.Sprintf(`---
title: "%s"
linkTitle: "%s"
description: "%s"
---

# %s

%s

## 🚀 Quick Start

`+"```bash"+`
# Get started with your project
make setup     # Install development tools
make check     # Verify everything works
%s%s`+"```"+`

## ✨ Features

This project provides a production-ready Go application with:

- **🏗️ Standard Project Layout** - Follows Go community best practices
- **🔧 Complete Toolchain** - golangci-lint, gofumpt, comprehensive quality gates
- **🧪 Progressive Testing** - Unit, integration, smoke%s tests
- **📦 Container Ready** - Multi-stage builds with distroless images (~10MB)
- **🚀 CI/CD Ready** - GitHub Actions for testing, security, and releases%s
- **⚡ Quality Gates** - Pre-commit hooks and comprehensive checks

## 🏛️ Architecture

%s

## 🎯 Getting Started

1. **Install Development Tools**
   `+"```bash"+`
   make setup
   `+"```"+`

2. **Run Quality Checks**
   `+"```bash"+`
   make check
   `+"```"+`

3. **Build and Run**
   `+"```bash"+`
   make build
%s   `+"```"+`

## 🛠️ Development

This project follows the **MODEST** principles for maintainable Go code:

- **Modularity** - Reusable, swappable components
- **Orthogonality** - Independent components with localized changes
- **Dependency Injection** - External dependencies passed explicitly
- **Explicitness** - Clear intent without magic
- **Single Responsibility** - Each component has one reason to change
- **Testability** - Designed for easy testing

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Run `+"```make check```"+` to ensure quality
4. Submit a pull request

## 📝 License

This project is available under the %s License.
`,
		config.ProjectName,
		config.ProjectName,
		config.Description,
		config.ProjectName,
		config.Description,
		generateRunCommands(config),
		"",
		generateTestingFeatures(config),
		generateDocumentationFeature(config),
		generateComponentDescription(config),
		generateRunCommands(config),
		config.License,
	)
}

func updateGettingStartedMarkdown(config *ProjectConfig) string {
	return fmt.Sprintf(`---
title: "Getting Started"
linkTitle: "Getting Started"
weight: 1
description: "Complete setup guide for %s"
---

# Getting Started

This guide will help you set up and start developing with %s.

## Prerequisites

Ensure you have the following installed:

- **Go 1.23+** - [Install from golang.org](https://golang.org/downloads/)
- **Git** - For version control
- **Make** - For build automation
- **Docker** (optional) - For containerized development

`+"```bash"+`
# Verify installations
go version    # Should show Go 1.23+
git --version
make --version
docker --version
`+"```"+`

## Development Workflow

### Daily Development

1. **Make changes** to your code
2. **Run quality checks** frequently:
   `+"```bash"+`
   make fmt    # Format code
   make test   # Run tests
   make lint   # Check for issues
   `+"```"+`
3. **Build and test** your applications:
   `+"```bash"+`
   make build
%s   `+"```"+`

### Quality Gates

The project enforces quality through multiple gates:

#### Pre-commit Hooks
Automatically installed during setup:
- Code formatting validation
- Basic linting checks
- Test execution

#### Comprehensive Checks
Run the full quality gate before commits:
`+"```bash"+`
make check
`+"```"+`

## Building Applications

### Local Development

`+"```bash"+`
# Build all applications
make build

# Run specific applications
%s
# Cross-platform builds
make build-all
`+"```"+`

### Docker Development

`+"```bash"+`
# Build Docker image
make docker-build

# Run in container
make docker-run
`+"```"+`

%s

## Next Steps

Now that your project is set up:

1. **Review the project structure** and customize as needed
2. **Start implementing your business logic**
3. **Add tests** for new functionality
4. **Update documentation** as you add features

## Troubleshooting

### Common Issues

**Go modules not working:**
`+"```bash"+`
go mod tidy
go mod download
`+"```"+`

**Linter failures:**
`+"```bash"+`
make fmt      # Fix formatting
make lint     # See specific issues
`+"```"+`

**Test failures:**
`+"```bash"+`
make test-unit     # Run unit tests only
go test -v ./...   # Verbose test output
`+"```"+`

**Docker build issues:**
`+"```bash"+`
make clean         # Clean build artifacts
make docker-build  # Rebuild image
`+"```"+`

### Getting Help

- Review the [README.md](../) for project overview
- Check the codebase for examples and patterns
- File issues on your project repository for bugs or questions
`,
		config.ProjectName,
		config.ProjectName,
		generateRunCommands(config),
		generateRunCommands(config),
		generateTestingSection(config),
	)
}

// Helper functions for dynamic content generation

func generateComponentDescription(config *ProjectConfig) string {
	components := []string{}
	if config.EnableCLI {
		components = append(components, "**CLI Application** - Command-line interface with flags and subcommands")
	}
	if config.EnableServer {
		components = append(components, "**HTTP Server** - REST API with graceful shutdown and health checks")
	}
	if config.EnableWorker {
		components = append(components, "**Background Worker** - Long-running process with signal handling")
	}

	if len(components) == 0 {
		return "This project provides a foundation for building Go applications with clean architecture patterns."
	}

	if len(components) == 1 {
		return fmt.Sprintf("This project includes:\n\n- %s", components[0])
	}

	result := "This project includes multiple components:\n\n"
	for _, component := range components {
		result += fmt.Sprintf("- %s\n", component)
	}
	return result
}

func generateRunCommands(config *ProjectConfig) string {
	commands := []string{}
	if config.EnableCLI {
		commands = append(commands, "make run-cli      # Run CLI application")
	}
	if config.EnableServer {
		commands = append(commands, "make run-server   # Run HTTP server")
	}
	if config.EnableWorker {
		commands = append(commands, "make run-worker   # Run background worker")
	}

	if len(commands) == 0 {
		return "go run ./..."
	}

	result := ""
	for _, cmd := range commands {
		result += fmt.Sprintf("   %s\n", cmd)
	}
	return result
}

func generateTestingFeatures(config *ProjectConfig) string {
	if config.EnableE2ETests {
		return ", and E2E"
	}
	return ""
}

func generateDocumentationFeature(config *ProjectConfig) string {
	if config.EnableDocs {
		return "\n- **📚 Documentation** - Hugo-powered static site with auto-generated API docs"
	}
	return ""
}

func generateTestingSection(config *ProjectConfig) string {
	if !config.EnableE2ETests {
		return ""
	}

	return `## Testing Strategy

The project implements a progressive testing approach:

### Test Categories

` + "```bash" + `
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
` + "```" + `

`
}

func generateReadme(config *ProjectConfig) error {
	readmeTemplate := `# {{.ProjectName}}

> {{.Description}}

Built from the [go-template-project](https://github.com/your-org/go-template-project)
starter template - providing quality gates, container deployment, and CI/CD without setup time.

## Quick Start

` + "```bash" + `
git clone {{.GitRemote}}
cd {{.ProjectName}}
make setup     # Install development tools
make check     # Verify everything works
{{if .EnableCLI}}go run ./cmd/cli{{end}}
{{if .EnableServer}}go run ./cmd/server{{end}}
` + "```" + `

## What You Get

**Zero-config development environment:**
- Formatting and linting that matches your CI
- Pre-commit hooks prevent broken commits
- Test coverage tracking with Codecov
- Security scanning built in

**Container-ready deployment:**
- Multi-stage Dockerfile → ~10MB images
- Distroless runtime for security
- Cross-platform builds
- Docker Compose for local development

**Quality automation:**
- GitHub Actions CI/CD
- Automated dependency updates
- Vulnerability scanning
- Code coverage enforcement (80% minimum)

## Available Commands

| Component | Command | Description |
|-----------|---------|-------------|
{{if .EnableCLI}}| CLI | ` + "`make run-cli`" + ` | Run command-line application |{{end}}
{{if .EnableServer}}| Server | ` + "`make run-server`" + ` | Run HTTP server on :8080 |{{end}}
{{if .EnableWorker}}| Worker | ` + "`make run-worker`" + ` | Run background worker |{{end}}
| All | ` + "`make build`" + ` | Build all binaries |
| Quality | ` + "`make check`" + ` | Run all quality checks |
{{if .EnableE2ETests}}| E2E Tests | ` + "`make test-e2e`" + ` | Run end-to-end tests |{{end}}
| Docker | ` + "`make docker-dev`" + ` | Start with docker-compose |

## Project Structure

` + "```" + `
{{.ProjectName}}/
├── cmd/                     # Application entry points
{{if .EnableCLI}}│   ├── cli/                 # Command-line interface{{end}}
{{if .EnableServer}}│   ├── server/             # HTTP server{{end}}
{{if .EnableWorker}}│   └── worker/             # Background worker{{end}}
├── internal/                # Private application code
│   ├── app/                 # Core business logic
│   ├── config/              # Configuration management
{{if .EnableServer}}│   └── handlers/           # HTTP handlers{{end}}
├── .github/workflows/       # CI/CD pipelines
├── docker/                  # Container configuration
├── scripts/                 # Development scripts
└── docs/                    # Documentation
` + "```" + `

## Development Workflow

1. **Make changes** - Edit code in ` + "`internal/`" + ` or add new commands in ` + "`cmd/`" + `
2. **Test locally** - ` + "`make check`" + ` runs all quality gates
3. **Commit** - Pre-commit hooks ensure consistency
4. **Push** - CI validates and deploys

## Configuration

Configure via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
{{if .EnableServer}}| ` + "`PORT`" + ` | ` + "`8080`" + ` | HTTP server port |{{end}}
{{if .EnableServer}}| ` + "`HOST`" + ` | ` + "`0.0.0.0`" + ` | HTTP server host |{{end}}
| ` + "`DEBUG`" + ` | ` + "`false`" + ` | Enable debug logging |
| ` + "`DATABASE_URL`" + ` | | Database connection string |

{{if .EnableServer}}## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| ` + "`/health`" + ` | GET | Health check |
| ` + "`/ready`" + ` | GET | Readiness check |
| ` + "`/api/info`" + ` | GET | Application info |
{{end}}

## License

{{.License}} - see LICENSE file for details.

---

*Generated from [go-template-project](https://github.com/your-org/go-template-project) -
A batteries-included Go starter template.*
`

	tmpl, err := template.New("readme").Parse(readmeTemplate)
	if err != nil {
		return err
	}

	file, err := os.Create("README.md")
	if err != nil {
		return err
	}
	defer file.Close()

	data := TemplateData{
		ProjectConfig: *config,
		Year:          "2024",
	}

	return tmpl.Execute(file, data)
}

func initializeGit(config *ProjectConfig) error {
	// Initialize git repository
	cmd := exec.Command("git", "init")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to initialize git: %w (output: %s)", err, string(output))
	}

	// Ensure git user config exists for commit (needed for E2E tests)
	if err := exec.Command("git", "config", "user.name", config.Author).Run(); err != nil {
		fmt.Printf("⚠️  Failed to set git user.name: %v\n", err)
	}

	if err := exec.Command("git", "config", "user.email", config.Email).Run(); err != nil {
		fmt.Printf("⚠️  Failed to set git user.email: %v\n", err)
	}

	// Add git remote if provided
	if config.GitRemote != "" {
		if err := exec.Command("git", "remote", "add", "origin", config.GitRemote).Run(); err != nil {
			fmt.Printf("⚠️  Failed to add git remote: %v\n", err)
		}
	}

	// Initial commit with timeout and better error handling
	addCmd := exec.Command("git", "add", ".")
	if output, err := addCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to stage files: %w (output: %s)", err, string(output))
	}

	// Use properly formatted commit message that passes pre-commit hooks
	commitMsg := fmt.Sprintf("feat: initialize %s project\n\nGenerated from go-template-project", config.ProjectName)
	commitCmd := exec.Command("git", "commit", "-m", commitMsg)

	// Set a timeout for the git commit to prevent hanging
	done := make(chan error, 1)
	go func() {
		output, err := commitCmd.CombinedOutput()
		if err != nil {
			done <- fmt.Errorf("failed to create initial commit: %w (output: %s)", err, string(output))
		} else {
			done <- nil
		}
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(10 * time.Second):
		if commitCmd.Process != nil {
			if err := commitCmd.Process.Kill(); err != nil {
				// Log kill error but don't fail the timeout handling
				fmt.Printf("Warning: failed to kill git commit process: %v\n", err)
			}
		}
		return fmt.Errorf("git commit timed out after 10 seconds")
	}
}

func setupPreCommitHooks() error {
	// Check if pre-commit is installed
	if err := exec.Command("pre-commit", "--version").Run(); err != nil {
		return fmt.Errorf("pre-commit not installed")
	}

	// Install hooks
	return exec.Command("pre-commit", "install").Run()
}

// Helper functions

func prompt(reader *bufio.Reader, question string) string {
	fmt.Printf("%s: ", question)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return ""
	}
	return strings.TrimSpace(answer)
}

func promptWithDefault(reader *bufio.Reader, question, defaultValue string) string {
	fmt.Printf("%s [%s]: ", question, defaultValue)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return defaultValue
	}
	answer = strings.TrimSpace(answer)
	if answer == "" {
		return defaultValue
	}
	return answer
}

func promptBool(reader *bufio.Reader, question string, defaultValue bool) bool {
	defaultStr := "y/N"
	if defaultValue {
		defaultStr = "Y/n"
	}

	fmt.Printf("%s [%s]: ", question, defaultStr)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return defaultValue
	}
	answer = strings.TrimSpace(strings.ToLower(answer))

	if answer == "" {
		return defaultValue
	}
	return answer == "y" || answer == "yes"
}

func isValidProjectName(name string) bool {
	matched, err := regexp.MatchString(projectNamePattern, name)
	if err != nil {
		return false
	}
	return matched && len(name) > 0
}

func isValidModulePath(path string) bool {
	matched, err := regexp.MatchString(modulePathPattern, path)
	if err != nil {
		return false
	}
	return matched
}

func getGitConfig(key, fallback string) string {
	cmd := exec.Command("git", "config", "--global", key)
	output, err := cmd.Output()
	if err != nil {
		return fallback
	}
	return strings.TrimSpace(string(output))
}
