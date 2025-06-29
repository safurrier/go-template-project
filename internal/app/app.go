package app

import (
	"fmt"
	"log"
	"os"
)

// App represents the core application.
type App struct {
	Name    string
	Version string
	Debug   bool
}

// New creates a new application instance.
func New(name, version string) *App {
	return &App{
		Name:    name,
		Version: version,
		Debug:   os.Getenv("DEBUG") == "true",
	}
}

// Run is the main entry point for CLI applications.
// Separated from main() to make testing easier.
func (a *App) Run() error {
	if a.Debug {
		log.Printf("Starting %s v%s in debug mode", a.Name, a.Version)
	}

	fmt.Printf("🚀 Hello from %s!\n", a.Name)
	fmt.Printf("   Version: %s\n", a.Version)
	
	if a.Debug {
		fmt.Println("   Debug mode: enabled")
	}

	return nil
}

// GetInfo returns basic application information.
func (a *App) GetInfo() map[string]string {
	return map[string]string{
		"name":    a.Name,
		"version": a.Version,
		"debug":   fmt.Sprintf("%t", a.Debug),
	}
}