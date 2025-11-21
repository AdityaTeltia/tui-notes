package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ssh-notes/terminal-notes/models"
)

// CLI provides command-line interface for power users
func runCLI() {
	exportCmd := flag.NewFlagSet("export", flag.ExitOnError)
	exportFormat := exportCmd.String("format", "markdown", "Export format: markdown, json, tar, zip")
	exportOutput := exportCmd.String("output", "./export", "Output path")
	exportUser := exportCmd.String("user", "", "Username")
	exportDataDir := exportCmd.String("data", "./data", "Data directory")

	importCmd := flag.NewFlagSet("import", flag.ExitOnError)
	importFormat := importCmd.String("format", "markdown", "Import format: markdown, json")
	importInput := importCmd.String("input", "", "Input path")
	importUser := importCmd.String("user", "", "Username")
	importDataDir := importCmd.String("data", "./data", "Data directory")

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "export":
		exportCmd.Parse(os.Args[2:])
		if *exportUser == "" {
			fmt.Println("Error: -user is required")
			os.Exit(1)
		}
		userDataDir := filepath.Join(*exportDataDir, *exportUser)
		model := models.NewMainModel(*exportUser, userDataDir)
		if err := model.ExportNotes(*exportFormat, *exportOutput); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Exported notes to %s\n", *exportOutput)

	case "import":
		importCmd.Parse(os.Args[2:])
		if *importUser == "" {
			fmt.Println("Error: -user is required")
			os.Exit(1)
		}
		if *importInput == "" {
			fmt.Println("Error: -input is required")
			os.Exit(1)
		}
		userDataDir := filepath.Join(*importDataDir, *importUser)
		os.MkdirAll(userDataDir, 0700)
		model := models.NewMainModel(*importUser, userDataDir)
		if err := model.ImportNotes(*importFormat, *importInput); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Imported notes from %s\n", *importInput)

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Terminal Notes CLI")
	fmt.Println("\nUsage:")
	fmt.Println("  ssh-notes export -user <username> -format <format> -output <path>")
	fmt.Println("  ssh-notes import -user <username> -format <format> -input <path>")
	fmt.Println("\nFormats:")
	fmt.Println("  export: markdown, json, tar, zip")
	fmt.Println("  import: markdown, json")
}

