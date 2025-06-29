package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/your-org/go-template-project/internal/config"
)

const (
	appName    = "go-template-worker"
	appVersion = "1.0.0"
)

// Worker represents a background worker.
type Worker struct {
	config *config.Config
	quit   chan bool
}

// NewWorker creates a new worker instance.
func NewWorker(cfg *config.Config) *Worker {
	return &Worker{
		config: cfg,
		quit:   make(chan bool),
	}
}

// Start begins the worker processing loop.
func (w *Worker) Start(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	log.Printf("🚀 Worker %s v%s started", appName, appVersion)

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Worker context cancelled")
			return
		case <-w.quit:
			log.Println("🛑 Worker quit signal received")
			return
		case <-ticker.C:
			w.processTask()
		}
	}
}

// Stop gracefully stops the worker.
func (w *Worker) Stop() {
	close(w.quit)
}

// processTask simulates processing a background task.
func (w *Worker) processTask() {
	if w.config.Debug {
		log.Println("📋 Processing task...")
	}
	
	// Simulate work
	time.Sleep(100 * time.Millisecond)
	
	if w.config.Debug {
		log.Println("✅ Task completed")
	}
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	worker := NewWorker(cfg)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start worker in goroutine
	go worker.Start(ctx)

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("🛑 Shutting down worker...")

	// Stop worker gracefully
	worker.Stop()
	cancel()

	// Give worker time to finish current task
	time.Sleep(2 * time.Second)

	log.Println("✅ Worker shut down gracefully")
}