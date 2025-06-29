// +build e2e

package e2e

import (
	"os"
	"os/exec"
	"testing"
	"time"
)

// TestCLIApplicationLaunches tests that the CLI application can start successfully.
// This is a smoke test that validates the complete CLI user journey.
func TestCLIApplicationLaunches(t *testing.T) {
	t.Parallel()

	// Arrange: Prepare CLI command
	cmd := exec.Command("go", "run", "./cmd/cli")
	cmd.Dir = getProjectRoot(t)

	// Act: Run CLI with timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	// Assert: CLI should exit successfully within reasonable time
	select {
	case err := <-done:
		// CLI exited - this is expected for a simple CLI
		if err != nil {
			// Check if it's a normal exit or actual error
			if exitError, ok := err.(*exec.ExitError); ok {
				if exitError.ExitCode() != 0 {
					t.Fatalf("CLI exited with non-zero code: %d", exitError.ExitCode())
				}
			}
		}
		// Success: CLI ran and exited normally
	case <-time.After(10 * time.Second):
		// If CLI is still running after 10 seconds, kill it
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		t.Fatal("CLI did not exit within 10 seconds - may be hanging")
	}
}

// TestCLIVersionFlag tests that the CLI responds to --version flag.
func TestCLIVersionFlag(t *testing.T) {
	t.Parallel()

	// Arrange: Prepare CLI command with --version
	cmd := exec.Command("go", "run", "./cmd/cli", "--version")
	cmd.Dir = getProjectRoot(t)

	// Act: Execute command
	output, err := cmd.CombinedOutput()

	// Assert: Should exit successfully and show version
	if err != nil {
		t.Fatalf("CLI --version failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if outputStr == "" {
		t.Fatal("CLI --version produced no output")
	}

	// Should contain app name and version info
	if !containsAppInfo(outputStr) {
		t.Fatalf("CLI --version output doesn't contain expected app info: %s", outputStr)
	}
}

// TestCLIHelp tests that the CLI provides help information.
func TestCLIHelp(t *testing.T) {
	t.Parallel()

	// Test both -h and --help flags
	testCases := []string{"-h", "--help"}

	for _, flag := range testCases {
		t.Run("flag_"+flag, func(t *testing.T) {
			// Arrange: Prepare CLI command with help flag
			cmd := exec.Command("go", "run", "./cmd/cli", flag)
			cmd.Dir = getProjectRoot(t)

			// Act: Execute command
			output, err := cmd.CombinedOutput()

			// Assert: Should provide help (may exit with code 2 which is normal for help)
			outputStr := string(output)
			if outputStr == "" {
				t.Fatalf("CLI %s produced no output", flag)
			}
			
			// Note: err may be non-nil for help flags, which is normal
			_ = err // Help flags may exit with non-zero code

			// Help output should contain usage information
			if !containsHelpInfo(outputStr) {
				t.Fatalf("CLI %s output doesn't contain expected help info: %s", flag, outputStr)
			}
		})
	}
}

// TestCLIInvalidFlag tests that the CLI handles invalid flags gracefully.
func TestCLIInvalidFlag(t *testing.T) {
	t.Parallel()

	// Arrange: Prepare CLI command with invalid flag
	cmd := exec.Command("go", "run", "./cmd/cli", "--invalid-flag-that-does-not-exist")
	cmd.Dir = getProjectRoot(t)

	// Act: Execute command
	output, err := cmd.CombinedOutput()

	// Assert: Should exit with error and provide helpful message
	if err == nil {
		t.Fatal("CLI should have failed with invalid flag but succeeded")
	}

	outputStr := string(output)
	if outputStr == "" {
		t.Fatal("CLI with invalid flag produced no error output")
	}

	// Should indicate the flag is unknown and possibly suggest help
	if !containsErrorInfo(outputStr) {
		t.Fatalf("CLI invalid flag output doesn't contain expected error info: %s", outputStr)
	}
}

// Helper functions

func getProjectRoot(t *testing.T) string {
	// Navigate to project root from tests/e2e/
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// If we're in tests/e2e, go up two levels
	if dir := wd; len(dir) > 8 && dir[len(dir)-8:] == "tests/e2e" {
		return "../.."
	}

	// If we're already in project root, use current directory
	if _, err := os.Stat("go.mod"); err == nil {
		return "."
	}

	t.Fatal("Could not determine project root directory")
	return ""
}

func containsAppInfo(output string) bool {
	// Look for version information in output
	// This is intentionally flexible to not couple to specific strings
	return len(output) > 10 && (
		// Common patterns for version output
		contains(output, "version") ||
		contains(output, "go-template") ||
		contains(output, "v1.") ||
		contains(output, "1.0"))
}

func containsHelpInfo(output string) bool {
	// Look for help patterns in output
	return len(output) > 10 && (
		contains(output, "Usage") ||
		contains(output, "usage") ||
		contains(output, "help") ||
		contains(output, "flag") ||
		contains(output, "option"))
}

func containsErrorInfo(output string) bool {
	// Look for error patterns in output
	return len(output) > 5 && (
		contains(output, "unknown") ||
		contains(output, "invalid") ||
		contains(output, "error") ||
		contains(output, "flag provided but not defined"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}