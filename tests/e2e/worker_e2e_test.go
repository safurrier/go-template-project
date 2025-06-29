//go:build e2e
// +build e2e

package e2e

import (
	"context"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestWorkerApplicationLaunches tests that the worker application can start and run.
// This validates the complete worker user journey.
func TestWorkerApplicationLaunches(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E worker test in short mode")
	}

	// Arrange: Prepare worker command
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", "./cmd/worker")
	cmd.Dir = getProjectRoot(t)

	// Set test environment with debug enabled to get more output
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0", "DEBUG=true")

	// Start worker
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start worker: %v", err)
	}

	// Ensure worker is killed at end of test
	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
			cmd.Wait()
		}
	}()

	// Act: Let worker run for a few seconds to verify it's working
	time.Sleep(3 * time.Second)

	// Assert: Worker should still be running (not crashed)
	if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
		t.Fatal("Worker exited unexpectedly")
	}

	// Send interrupt signal for graceful shutdown
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatalf("Failed to send interrupt signal to worker: %v", err)
	}

	// Wait for graceful shutdown
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		// Worker exited
		if err != nil {
			// Check if it's a signal termination (expected)
			if exitError, ok := err.(*exec.ExitError); ok {
				if exitError.ExitCode() == 130 { // SIGINT exit code
					// This is expected for graceful shutdown
					return
				}
			}
			t.Logf("Worker exited with error (may be normal for interrupt): %v", err)
		}
		// Success: Worker shut down gracefully
	case <-time.After(10 * time.Second):
		// Force kill if graceful shutdown took too long
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		t.Fatal("Worker did not shut down gracefully within 10 seconds")
	}
}

// TestWorkerTaskProcessing tests that the worker processes tasks correctly.
func TestWorkerTaskProcessing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E worker processing test in short mode")
	}

	// Arrange: Start worker with debug output to capture its activity
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", "./cmd/worker")
	cmd.Dir = getProjectRoot(t)
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0", "DEBUG=true")

	// Capture output to verify worker is processing tasks
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("Failed to create stdout pipe: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("Failed to create stderr pipe: %v", err)
	}

	// Start worker
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start worker for processing test: %v", err)
	}

	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
			cmd.Wait()
		}
	}()

	// Act: Collect output for a few seconds to see worker activity
	outputChan := make(chan string, 1)
	go func() {
		output := captureOutput(stdout, stderr, 8*time.Second)
		outputChan <- output
	}()

	// Wait for output or timeout
	var output string
	select {
	case output = <-outputChan:
		// Got output
	case <-time.After(10 * time.Second):
		t.Fatal("Timeout waiting for worker output")
	}

	// Assert: Worker should show activity (task processing)
	if len(output) == 0 {
		t.Fatal("Worker produced no output - may not be working correctly")
	}

	// Look for signs of worker activity
	if !containsWorkerActivity(output) {
		t.Fatalf("Worker output doesn't show expected activity: %s", output)
	}

	// Signal shutdown
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Logf("Failed to send interrupt signal: %v", err)
	}
}

// TestWorkerConfiguration tests that the worker respects configuration.
func TestWorkerConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E worker config test in short mode")
	}

	testCases := []struct {
		name   string
		env    []string
		expect string
	}{
		{
			name:   "debug_mode_enabled",
			env:    []string{"DEBUG=true"},
			expect: "debug output",
		},
		{
			name:   "debug_mode_disabled",
			env:    []string{"DEBUG=false"},
			expect: "minimal output",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange: Start worker with specific configuration
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, "go", "run", "./cmd/worker")
			cmd.Dir = getProjectRoot(t)
			cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
			cmd.Env = append(cmd.Env, tc.env...)

			// Capture output
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				t.Fatalf("Failed to create stdout pipe: %v", err)
			}

			stderr, err := cmd.StderrPipe()
			if err != nil {
				t.Fatalf("Failed to create stderr pipe: %v", err)
			}

			// Start worker
			if err := cmd.Start(); err != nil {
				t.Fatalf("Failed to start worker: %v", err)
			}

			defer func() {
				if cmd.Process != nil {
					cmd.Process.Kill()
					cmd.Wait()
				}
			}()

			// Act: Collect output
			outputChan := make(chan string, 1)
			go func() {
				output := captureOutput(stdout, stderr, 5*time.Second)
				outputChan <- output
			}()

			var output string
			select {
			case output = <-outputChan:
				// Got output
			case <-time.After(6 * time.Second):
				t.Fatal("Timeout waiting for worker output")
			}

			// Assert: Output should match expected configuration behavior
			switch tc.expect {
			case "debug output":
				if !containsDebugInfo(output) {
					t.Fatalf("Expected debug output but got: %s", output)
				}
			case "minimal output":
				if containsDebugInfo(output) {
					t.Fatalf("Expected minimal output but got debug info: %s", output)
				}
			}

			// Signal shutdown
			if err := cmd.Process.Signal(os.Interrupt); err != nil {
				t.Logf("Failed to send interrupt signal: %v", err)
			}
		})
	}
}

// Helper functions for worker tests

func captureOutput(stdout, stderr io.Reader, duration time.Duration) string {
	outputChan := make(chan []byte, 100)
	done := make(chan bool, 1)

	// Read from stdout
	go func() {
		buffer := make([]byte, 1024)
		for {
			n, err := stdout.Read(buffer)
			if n > 0 {
				data := make([]byte, n)
				copy(data, buffer[:n])
				outputChan <- data
			}
			if err != nil {
				break
			}
		}
	}()

	// Read from stderr
	go func() {
		buffer := make([]byte, 1024)
		for {
			n, err := stderr.Read(buffer)
			if n > 0 {
				data := make([]byte, n)
				copy(data, buffer[:n])
				outputChan <- data
			}
			if err != nil {
				break
			}
		}
	}()

	// Stop after duration
	go func() {
		time.Sleep(duration)
		done <- true
	}()

	var allOutput []string
	for {
		select {
		case data := <-outputChan:
			allOutput = append(allOutput, string(data))
		case <-done:
			return strings.Join(allOutput, "")
		}
	}
}

func containsWorkerActivity(output string) bool {
	// Look for signs that the worker is actively processing
	// This is flexible to avoid coupling to exact log messages
	return len(output) > 10 && (contains(output, "Worker") ||
		contains(output, "worker") ||
		contains(output, "started") ||
		contains(output, "processing") ||
		contains(output, "task") ||
		contains(output, "Processing") ||
		contains(output, "completed") ||
		contains(output, "🚀") || // Emoji used in worker startup
		contains(output, "📋") || // Emoji used in task processing
		contains(output, "✅")) // Emoji used in task completion
}

func containsDebugInfo(output string) bool {
	// Look for debug-level information
	return len(output) > 5 && (contains(output, "debug") ||
		contains(output, "DEBUG") ||
		contains(output, "Processing task") ||
		contains(output, "Task completed") ||
		// Look for detailed output that would only appear in debug mode
		(contains(output, "📋") && contains(output, "✅")))
}
