package main

import (
	"flag"
	"log"
	"os"

	"github.com/your-org/go-template-project/internal/app"
)

const (
	appName    = "go-template-cli"
	appVersion = "1.0.0"
)

func main() {
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		log.Printf("%s version %s", appName, appVersion)
		os.Exit(0)
	}

	application := app.New(appName, appVersion)

	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
