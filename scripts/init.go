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
