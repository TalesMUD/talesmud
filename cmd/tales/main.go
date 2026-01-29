package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/joho/godotenv"
	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/importer"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/server"
)

func main() {
	// Parse command-line flags
	importFolder := flag.String("import", "", "Import world data from folder (e.g., mvp-rpg-1)")
	verbose := flag.Bool("verbose", false, "Enable verbose output during import")
	dryRun := flag.Bool("dry-run", false, "Validate import data without making changes")
	flag.Parse()

	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Warn("Error loading .env file")
	}

	// Configure logging
	logLevel := strings.ToLower(strings.TrimSpace(os.Getenv("LOG_LEVEL")))
	if logLevel == "" {
		logLevel = "info"
	}
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		log.WithField("LOG_LEVEL", logLevel).Warn("Invalid LOG_LEVEL, defaulting to info")
		level = log.InfoLevel
	}
	log.SetLevel(level)

	// Get SQLite path
	sqlitePath := os.Getenv("SQLITE_PATH")
	if sqlitePath == "" {
		sqlitePath = "talesmud.db"
	}

	// Handle import command
	if *importFolder != "" {
		runImport(*importFolder, sqlitePath, *verbose, *dryRun)
		return
	}

	// Start the server
	fmt.Println("Starting tales server...")
	fmt.Printf("SQLite database: %v\n", sqlitePath)

	srv := server.NewApp()
	srv.Run()
}

func runImport(folderName, sqlitePath string, verbose, dryRun bool) {
	importPath := filepath.Join("import", folderName)

	// Validate import folder exists
	if _, err := os.Stat(importPath); os.IsNotExist(err) {
		log.Fatalf("Import folder not found: %s", importPath)
	}

	fmt.Println("===========================================")
	fmt.Println("TalesMUD World Importer")
	fmt.Println("===========================================")
	fmt.Printf("Import folder: %s\n", importPath)
	fmt.Printf("Database: %s\n", sqlitePath)
	fmt.Printf("Verbose: %v\n", verbose)
	fmt.Printf("Dry-run: %v\n", dryRun)
	fmt.Println("-------------------------------------------")

	// Initialize database
	client, err := dbsqlite.Open(sqlitePath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer client.Close()

	repos := repository.NewSQLiteFactory(client)

	// Create and run importer
	imp := importer.New(repos, importPath)
	imp.SetVerbose(verbose)
	imp.SetDryRun(dryRun)

	result, err := imp.Import()
	if err != nil {
		log.Fatalf("Import failed: %v", err)
	}

	// Print results
	fmt.Println("-------------------------------------------")
	fmt.Println("Import Results:")
	fmt.Printf("  Scripts:     %d\n", result.ScriptsImported)
	fmt.Printf("  Items:       %d\n", result.ItemsImported)
	fmt.Printf("  Loot Tables: %d\n", result.LootTablesImported)
	fmt.Printf("  NPCs:        %d\n", result.NPCsImported)
	fmt.Printf("  Dialogs:     %d\n", result.DialogsImported)
	fmt.Printf("  Rooms:       %d\n", result.RoomsImported)
	fmt.Printf("  Assets:      %d\n", result.AssetsImported)
	fmt.Printf("  Characters:  %d relocated\n", result.CharactersRelocated)
	fmt.Printf("  Duration:    %v\n", result.Duration)

	if result.Backup != "" {
		fmt.Printf("  Backup:      %s\n", result.Backup)
	}

	if len(result.Errors) > 0 {
		fmt.Println("-------------------------------------------")
		fmt.Printf("Errors (%d):\n", len(result.Errors))
		for _, e := range result.Errors {
			fmt.Printf("  - %s\n", e)
		}
	}

	fmt.Println("===========================================")
	if dryRun {
		fmt.Println("Dry-run complete. No changes were made.")
	} else if len(result.Errors) == 0 {
		fmt.Println("Import completed successfully!")
	} else {
		fmt.Println("Import completed with errors.")
		os.Exit(1)
	}
}
