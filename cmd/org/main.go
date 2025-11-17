package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/rwejlgaard/org/internal/config"
	"github.com/rwejlgaard/org/internal/model"
	"github.com/rwejlgaard/org/internal/parser"
	"github.com/rwejlgaard/org/internal/ui"
)

func main() {
	var filePath string
	var multiMode bool
	var captureMode bool
	flag.BoolVar(&multiMode, "multi", false, "Load all org files in current directory as top-level items")
	flag.BoolVar(&multiMode, "m", false, "Load all org files in current directory (shorthand)")
	flag.BoolVar(&captureMode, "capture", false, "Start in capture mode")
	flag.BoolVar(&captureMode, "c", false, "Start in capture mode (shorthand)")
	flag.Parse()

	// Check for positional argument or capture text
	var captureText string
	if len(flag.Args()) > 0 {
		if captureMode {
			// First argument is capture text when in capture mode
			captureText = flag.Args()[0]
			// Second argument (if present) is the file path
			if len(flag.Args()) > 1 {
				filePath = flag.Args()[1]
			}
		} else {
			// First argument is file path in normal mode
			filePath = flag.Args()[0]
		}
	}

	// Check if input is being piped
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// Data is being piped to stdin
		reader := bufio.NewReader(os.Stdin)
		pipedText, err := io.ReadAll(reader)
		if err == nil && len(pipedText) > 0 {
			captureMode = true
			captureText = string(pipedText)
			// If no file path was provided via args, check if last arg could be a path
			if filePath == "" && len(flag.Args()) > 0 {
				filePath = flag.Args()[0]
			}
		}
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Error loading config, using defaults: %v\n", err)
		cfg = config.DefaultConfig()
	}

	var orgFile *model.OrgFile

	if multiMode {
		// Multi-file mode: load all .org files in directory
		var dirPath string
		if filePath != "" {
			// Check if provided path is a directory
			info, err := os.Stat(filePath)
			if err == nil && info.IsDir() {
				dirPath = filePath
			} else {
				// Use directory of the provided file path
				dirPath = filepath.Dir(filePath)
			}
		} else {
			// Use current directory
			cwd, err := os.Getwd()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
				os.Exit(1)
			}
			dirPath = cwd
		}

		orgFile, err = parser.ParseMultipleOrgFiles(dirPath, cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing org files: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Single file mode (default)
		if filePath == "" {
			// Default to ./todo.org
			cwd, err := os.Getwd()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
				os.Exit(1)
			}
			filePath = filepath.Join(cwd, "todo.org")
		}

		// Parse the org file
		orgFile, err = parser.ParseOrgFile(filePath, cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing org file: %v\n", err)
			os.Exit(1)
		}
	}

	// Run the UI
	if err := ui.RunUI(orgFile, cfg, captureMode, captureText); err != nil {
		fmt.Fprintf(os.Stderr, "Error running UI: %v\n", err)
		os.Exit(1)
	}

	// Save on exit
	if err := parser.Save(orgFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving file: %v\n", err)
		os.Exit(1)
	}
}
