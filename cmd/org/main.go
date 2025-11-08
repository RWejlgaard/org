package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rwejlgaard/org/internal/config"
	"github.com/rwejlgaard/org/internal/parser"
	"github.com/rwejlgaard/org/internal/ui"
)

func main() {
	var filePath string
	flag.StringVar(&filePath, "file", "", "Path to org file (default: ./todo.org)")
	flag.StringVar(&filePath, "f", "", "Path to org file (shorthand)")
	flag.Parse()

	// Check for positional argument first
	if filePath == "" && len(flag.Args()) > 0 {
		filePath = flag.Args()[0]
	}

	// Default to ./todo.org if no file specified
	if filePath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
			os.Exit(1)
		}
		filePath = filepath.Join(cwd, "todo.org")
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Error loading config, using defaults: %v\n", err)
		cfg = config.DefaultConfig()
	}

	// Parse the org file
	orgFile, err := parser.ParseOrgFile(filePath, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing org file: %v\n", err)
		os.Exit(1)
	}

	// Run the UI
	if err := ui.RunUI(orgFile, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error running UI: %v\n", err)
		os.Exit(1)
	}

	// Save on exit
	if err := parser.Save(orgFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving file: %v\n", err)
		os.Exit(1)
	}
}
