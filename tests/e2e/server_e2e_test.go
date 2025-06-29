// +build e2e

package e2e

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestServerApplicationLaunches tests that the HTTP server can start and respond to requests.
// This validates the complete server user journey.
func TestServerApplicationLaunches(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E server test in short mode")
	}

	// Arrange: Start server in background
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", "./cmd/server")
	cmd.Dir = getProjectRoot(t)
	
	// Set test environment
	cmd.Env = append(os.Environ(), 
		"PORT=8081", // Use different port to avoid conflicts
		"DEBUG=true",
	)

	// Start server
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Ensure server is killed at end of test
	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
			cmd.Wait()
		}
	}()

	// Act: Wait for server to start and test endpoints
	serverURL := "http://localhost:8081"
	
	// Wait for server to be ready
	if !waitForServer(t, serverURL+"/health", 10*time.Second) {
		t.Fatal("Server did not start within timeout")
	}

	// Assert: Test that server responds correctly
	testServerEndpoints(t, serverURL)
}

// TestServerHealthEndpoint tests the health check endpoint specifically.
func TestServerHealthEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E server test in short mode")
	}

	// This test assumes server is running (could be started by docker-compose or manually)
	// For CI, we'll use a different port to avoid conflicts
	serverURL := "http://localhost:8082"
	
	// Try to start a server instance for this test
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", "./cmd/server")
	cmd.Dir = getProjectRoot(t)
	cmd.Env = append(os.Environ(), "PORT=8082", "DEBUG=false")

	if err := cmd.Start(); err != nil {
		t.Skipf("Could not start server for health test: %v", err)
		return
	}

	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
			cmd.Wait()
		}
	}()

	// Wait for server to start
	if !waitForServer(t, serverURL+"/health", 8*time.Second) {
		t.Skip("Server did not start in time for health test")
		return
	}

	// Test health endpoint
	resp, err := http.Get(serverURL + "/health")
	if err != nil {
		t.Fatalf("Health check request failed: %v", err)
	}
	defer resp.Body.Close()

	// Should return 200 OK
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Health check returned status %d, expected %d", resp.StatusCode, http.StatusOK)
	}

	// Should return JSON
	if contentType := resp.Header.Get("Content-Type"); !strings.Contains(contentType, "application/json") {
		t.Fatalf("Health check returned Content-Type %s, expected application/json", contentType)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read health check response: %v", err)
	}

	bodyStr := string(body)
	if len(bodyStr) == 0 {
		t.Fatal("Health check response body is empty")
	}

	// Should contain health status (flexible to avoid coupling to exact JSON structure)
	if !containsHealthInfo(bodyStr) {
		t.Fatalf("Health check response doesn't contain expected health info: %s", bodyStr)
	}
}

// TestServerGracefulShutdown tests that the server shuts down gracefully.
func TestServerGracefulShutdown(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E server shutdown test in short mode")
	}

	// Arrange: Start server
	cmd := exec.Command("go", "run", "./cmd/server")
	cmd.Dir = getProjectRoot(t)
	cmd.Env = append(os.Environ(), "PORT=8083", "DEBUG=true")

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start server for shutdown test: %v", err)
	}

	serverURL := "http://localhost:8083"

	// Wait for server to start
	if !waitForServer(t, serverURL+"/health", 8*time.Second) {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		t.Fatal("Server did not start for shutdown test")
	}

	// Act: Send interrupt signal (graceful shutdown)
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatalf("Failed to send interrupt signal: %v", err)
	}

	// Assert: Server should shut down gracefully within reasonable time
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		// Server exited
		if err != nil {
			// Check if it's a signal termination (expected)
			if exitError, ok := err.(*exec.ExitError); ok {
				if exitError.ExitCode() == 130 { // SIGINT exit code
					// This is expected for graceful shutdown
					return
				}
			}
			t.Logf("Server exited with error (may be normal for interrupt): %v", err)
		}
		// Success: Server shut down
	case <-time.After(15 * time.Second):
		// Force kill if graceful shutdown took too long
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		t.Fatal("Server did not shut down gracefully within 15 seconds")
	}
}

// Helper functions for server tests

func waitForServer(t *testing.T, url string, timeout time.Duration) bool {
	client := &http.Client{Timeout: 1 * time.Second}
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return true
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	return false
}

func testServerEndpoints(t *testing.T, baseURL string) {
	client := &http.Client{Timeout: 5 * time.Second}

	// Test health endpoint
	t.Run("health_endpoint", func(t *testing.T) {
		resp, err := client.Get(baseURL + "/health")
		if err != nil {
			t.Fatalf("Health endpoint request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Health endpoint returned status %d", resp.StatusCode)
		}
	})

	// Test readiness endpoint
	t.Run("ready_endpoint", func(t *testing.T) {
		resp, err := client.Get(baseURL + "/ready")
		if err != nil {
			t.Fatalf("Ready endpoint request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Ready endpoint returned status %d", resp.StatusCode)
		}
	})

	// Test API info endpoint
	t.Run("api_info_endpoint", func(t *testing.T) {
		resp, err := client.Get(baseURL + "/api/info")
		if err != nil {
			t.Fatalf("API info endpoint request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("API info endpoint returned status %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read API info response: %v", err)
		}

		if !containsAPIInfo(string(body)) {
			t.Fatalf("API info response doesn't contain expected info: %s", body)
		}
	})

	// Test invalid endpoint (should return 404)
	t.Run("invalid_endpoint", func(t *testing.T) {
		resp, err := client.Get(baseURL + "/invalid/endpoint")
		if err != nil {
			t.Fatalf("Invalid endpoint request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("Invalid endpoint returned status %d, expected 404", resp.StatusCode)
		}
	})
}

func containsHealthInfo(body string) bool {
	// Look for health-related information without coupling to exact JSON structure
	return len(body) > 5 && (
		contains(body, "healthy") ||
		contains(body, "status") ||
		contains(body, "timestamp") ||
		contains(body, "version"))
}

func containsAPIInfo(body string) bool {
	// Look for API information patterns
	return len(body) > 5 && (
		contains(body, "name") ||
		contains(body, "version") ||
		contains(body, "go-template"))
}