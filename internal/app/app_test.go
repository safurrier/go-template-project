package app

import (
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	app := New("test-app", "1.0.0")
	
	if app.Name != "test-app" {
		t.Errorf("Expected name 'test-app', got '%s'", app.Name)
	}
	
	if app.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", app.Version)
	}
}

func TestRun(t *testing.T) {
	app := New("test-app", "1.0.0")
	
	err := app.Run()
	if err != nil {
		t.Errorf("Run() returned error: %v", err)
	}
}

func TestDebugMode(t *testing.T) {
	// Test debug mode enabled
	os.Setenv("DEBUG", "true")
	defer os.Unsetenv("DEBUG")
	
	app := New("test-app", "1.0.0")
	if !app.Debug {
		t.Error("Expected debug mode to be enabled")
	}
	
	// Test debug mode disabled
	os.Setenv("DEBUG", "false")
	app = New("test-app", "1.0.0")
	if app.Debug {
		t.Error("Expected debug mode to be disabled")
	}
}

func TestGetInfo(t *testing.T) {
	app := New("test-app", "1.0.0")
	info := app.GetInfo()
	
	expectedFields := []string{"name", "version", "debug"}
	for _, field := range expectedFields {
		if _, exists := info[field]; !exists {
			t.Errorf("Expected field '%s' in info", field)
		}
	}
	
	if info["name"] != "test-app" {
		t.Errorf("Expected name 'test-app', got '%s'", info["name"])
	}
	
	if info["version"] != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", info["version"])
	}
}