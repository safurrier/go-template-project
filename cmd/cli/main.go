package main

import (
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/your-org/go-template-project/internal/app"
)

const (
	appName    = "go-template-cli"
	appVersion = "1.0.0"
)

func main() {
	showVersion := flag.Bool("version", false, "Show version information")
	sessionKey := flag.Int("session", 0, "Load a specific session key (defaults to the latest session)")
	driver := flag.String("driver", "", "Focus on a driver by number, name, or acronym")
	refresh := flag.Duration("refresh", 5*time.Second, "Refresh interval for live data")
	baseURL := flag.String("api", "", "Override the OpenF1 API base URL")
	flag.Parse()

	if *showVersion {
		log.Printf("%s version %s", appName, appVersion)
		os.Exit(0)
	}

	options := []app.Option{
		app.WithRefreshInterval(*refresh),
	}
	if *sessionKey > 0 {
		options = append(options, app.WithSessionKey(*sessionKey))
	}
	if trimmed := strings.TrimSpace(*driver); trimmed != "" {
		options = append(options, app.WithDriverFilter(trimmed))
	}
	if trimmed := strings.TrimSpace(*baseURL); trimmed != "" {
		options = append(options, app.WithBaseURL(trimmed))
	}

	application := app.New(appName, appVersion, options...)

	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
