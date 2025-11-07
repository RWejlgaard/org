package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
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

	// Parse the org file
	orgFile, err := ParseOrgFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing org file: %v\n", err)
		os.Exit(1)
	}

	// Run the UI
	if err := runUI(orgFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error running UI: %v\n", err)
		os.Exit(1)
	}

	// Save on exit
	if err := orgFile.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving file: %v\n", err)
		os.Exit(1)
	}
}
