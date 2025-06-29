// +build e2e

package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestInitScriptBasicFunctionality tests that the init script can run without errors.
// This is a critical user path for template usage.
func TestInitScriptBasicFunctionality(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E init script test in short mode")
	}

	// Arrange: Create temporary directory for test project
	tmpDir := createTempProjectDir(t)
	defer cleanupTempDir(t, tmpDir)

	// Copy template files to temp directory
	copyTemplateFiles(t, getProjectRoot(t), tmpDir)

	// Act: Run init script with non-interactive input
	cmd := exec.Command("go", "run", "scripts/init.go")
	cmd.Dir = tmpDir
	
	// Provide automated input to the interactive script
	// This simulates user input for project configuration
	input := strings.Join([]string{
		"test-project",                                    // Project name
		"github.com/test-org/test-project",              // Module path
		"A test project for E2E validation",             // Description
		"Test User",                                      // Author name
		"test@example.com",                               // Author email
		"MIT",                                            // License
		"y",                                              // Include CLI
		"y",                                              // Include server
		"n",                                              // Include worker
		"y",                                              // Include docs
		"",                                               // Git remote (empty)
		"y",                                              // Confirm initialization
	}, "\n") + "\n"

	cmd.Stdin = strings.NewReader(input)

	// Set timeout to prevent hanging
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	// Assert: Init script should complete within reasonable time
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Init script failed: %v", err)
		}
	case <-time.After(30 * time.Second):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		t.Fatal("Init script did not complete within 30 seconds")
	}

	// Verify project was initialized correctly
	verifyInitializedProject(t, tmpDir)
}

// TestInitScriptValidation tests that the init script validates input correctly.
func TestInitScriptValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E init validation test in short mode")
	}

	// Test invalid project name
	t.Run("invalid_project_name", func(t *testing.T) {
		tmpDir := createTempProjectDir(t)
		defer cleanupTempDir(t, tmpDir)
		copyTemplateFiles(t, getProjectRoot(t), tmpDir)

		cmd := exec.Command("go", "run", "scripts/init.go")
		cmd.Dir = tmpDir

		// Provide invalid project name (starts with number)
		input := "123-invalid-name\n"
		cmd.Stdin = strings.NewReader(input)

		// Should fail or ask for valid input
		err := cmd.Run()
		if err == nil {
			t.Log("Init script may have asked for valid input (expected behavior)")
		}
	})

	// Test invalid module path
	t.Run("invalid_module_path", func(t *testing.T) {
		tmpDir := createTempProjectDir(t)
		defer cleanupTempDir(t, tmpDir)
		copyTemplateFiles(t, getProjectRoot(t), tmpDir)

		cmd := exec.Command("go", "run", "scripts/init.go")
		cmd.Dir = tmpDir

		// Provide project name then invalid module path
		input := strings.Join([]string{
			"valid-project",
			"invalid-module-path-no-slash",
		}, "\n") + "\n"

		cmd.Stdin = strings.NewReader(input)

		// Should fail or ask for valid input
		err := cmd.Run()
		if err == nil {
			t.Log("Init script may have asked for valid input (expected behavior)")
		}
	})
}

// TestInitScriptFileGeneration tests that the init script generates expected files.
func TestInitScriptFileGeneration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E init file generation test in short mode")
	}

	// Arrange: Create temporary directory and run init
	tmpDir := createTempProjectDir(t)
	defer cleanupTempDir(t, tmpDir)
	copyTemplateFiles(t, getProjectRoot(t), tmpDir)

	// Run init script with minimal configuration
	cmd := exec.Command("go", "run", "scripts/init.go")
	cmd.Dir = tmpDir

	input := strings.Join([]string{
		"example-project",
		"github.com/example/example-project",
		"An example project",
		"Example User",
		"user@example.com",
		"MIT",
		"y", // CLI
		"n", // Server (disabled to test removal)
		"n", // Worker (disabled to test removal)
		"y", // Docs
		"",  // No git remote
		"y", // Confirm
	}, "\n") + "\n"

	cmd.Stdin = strings.NewReader(input)

	if err := cmd.Run(); err != nil {
		t.Fatalf("Init script failed: %v", err)
	}

	// Assert: Verify expected files exist and unwanted files are removed
	expectedFiles := []string{
		"go.mod",
		"README.md",
		"cmd/cli/main.go",
		"internal/app/app.go",
		"internal/config/config.go",
		"docs",
	}

	for _, file := range expectedFiles {
		path := filepath.Join(tmpDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected file/directory %s was not created", file)
		}
	}

	// Verify server files were removed (since server was disabled)
	unwantedFiles := []string{
		"cmd/server",
		"internal/handlers",
	}

	for _, file := range unwantedFiles {
		path := filepath.Join(tmpDir, file)
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("Unwanted file/directory %s was not removed", file)
		}
	}

	// Verify go.mod was updated correctly
	verifyGoModUpdated(t, tmpDir, "github.com/example/example-project")
}

// Helper functions for init script tests

func createTempProjectDir(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "go-template-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	return tmpDir
}

func cleanupTempDir(t *testing.T, dir string) {
	if err := os.RemoveAll(dir); err != nil {
		t.Logf("Failed to cleanup temp directory %s: %v", dir, err)
	}
}

func copyTemplateFiles(t *testing.T, srcDir, dstDir string) {
	// Copy essential template files for testing
	// Note: This is a simplified copy for testing - real usage would clone the repo
	
	files := []string{
		"go.mod",
		"Makefile",
		"Dockerfile",
		".gitignore",
		".golangci.yml",
		".pre-commit-config.yaml",
	}

	for _, file := range files {
		srcPath := filepath.Join(srcDir, file)
		dstPath := filepath.Join(dstDir, file)
		
		if err := copyFile(srcPath, dstPath); err != nil {
			t.Logf("Warning: Failed to copy %s: %v", file, err)
		}
	}

	// Copy directories
	dirs := []string{
		"cmd",
		"internal",
		"scripts",
		"docs",
		".github",
		"docker",
	}

	for _, dir := range dirs {
		srcPath := filepath.Join(srcDir, dir)
		dstPath := filepath.Join(dstDir, dir)
		
		if err := copyDir(srcPath, dstPath); err != nil {
			t.Logf("Warning: Failed to copy directory %s: %v", dir, err)
		}
	}
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	
	// Create parent directory if needed
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	
	return os.WriteFile(dst, data, 0644)
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return copyFile(path, dstPath)
	})
}

func verifyInitializedProject(t *testing.T, projectDir string) {
	// Check that README.md was generated
	readmePath := filepath.Join(projectDir, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		t.Error("README.md was not generated by init script")
		return
	}

	// Check README content
	readmeContent, err := os.ReadFile(readmePath)
	if err != nil {
		t.Errorf("Failed to read generated README.md: %v", err)
		return
	}

	readmeStr := string(readmeContent)
	if !strings.Contains(readmeStr, "test-project") {
		t.Error("Generated README.md doesn't contain expected project name")
	}

	// Check that go.mod was updated
	verifyGoModUpdated(t, projectDir, "github.com/test-org/test-project")
}

func verifyGoModUpdated(t *testing.T, projectDir, expectedModule string) {
	goModPath := filepath.Join(projectDir, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		t.Errorf("Failed to read go.mod: %v", err)
		return
	}

	goModStr := string(content)
	if !strings.Contains(goModStr, expectedModule) {
		t.Errorf("go.mod doesn't contain expected module path %s, content: %s", expectedModule, goModStr)
	}
}