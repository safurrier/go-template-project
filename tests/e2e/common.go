//go:build e2e
// +build e2e

package e2e

import (
	"os"
	"strings"
	"testing"
)

// getProjectRoot finds the project root directory from any location within the project.
// It works by looking for the go.mod file starting from the current directory and going up.
func getProjectRoot(t *testing.T) string {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Try current directory first (CI might run from project root)
	if _, err := os.Stat("go.mod"); err == nil {
		return "."
	}

	// If we're in tests/e2e, go up two levels
	if dir := wd; len(dir) >= 8 && dir[len(dir)-8:] == "tests/e2e" {
		return "../.."
	}

	// Try going up one level (in case we're in tests/)
	if _, err := os.Stat("../go.mod"); err == nil {
		return ".."
	}

	// Try going up two levels (in case we're in tests/e2e/)
	if _, err := os.Stat("../../go.mod"); err == nil {
		return "../.."
	}

	t.Logf("Working directory: %s", wd)
	t.Fatal("Could not determine project root directory")
	return ""
}

// contains checks if a string contains a substring.
// This is a shared helper to avoid duplicating the logic across test files.
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
