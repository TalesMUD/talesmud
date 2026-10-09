package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/joho/godotenv"
	"github.com/talesmud/talesmud/pkg/classkit"
	"github.com/talesmud/talesmud/pkg/contenthealth"
	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/gamemode"
	"github.com/talesmud/talesmud/pkg/importer"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/server"
	"github.com/talesmud/talesmud/pkg/service"
	"github.com/talesmud/talesmud/pkg/util"
)

func main() {
	// Parse command-line flags
	importFolder := flag.String("import", "", "Import world data from folder (e.g., mvp-rpg-1)")
	checkFolder := flag.String("check", "", "Run content health against an import folder name")
	checkJSON := flag.Bool("json", false, "With -check, print the full report as JSON")
	failOn := flag.String("fail-on", "error", "With -check, exit 1 at this severity: error or warning")
	configPath := flag.String("config", "", "Game mode YAML. Sets port and database when those fields are present.")
	verbose := flag.Bool("verbose", false, "Enable verbose output during import")
	dryRun := flag.Bool("dry-run", false, "Validate import data without making changes")
	flag.Parse()
	if *checkFolder != "" && *importFolder != "" {
		log.Fatal("-check and -import cannot be used together")
	}

	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Warn("Error loading .env file")
	}
	if *configPath != "" {
		if err := gamemode.ApplyFile(*configPath); err != nil {
			log.Fatal(err)
		}
	}
	gamemode.ApplyEnv()

	// Configure logging (stderr + rotating file, 7-day retention)
	util.ConfigureLogging()

	// Get SQLite path
	sqlitePath := os.Getenv("SQLITE_PATH")
	if sqlitePath == "" {
		sqlitePath = "talesmud.db"
	}

	if *checkFolder != "" {
		os.Exit(runCheck(*checkFolder, *checkJSON, *failOn))
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
	fmt.Printf("  Spawners:    %d\n", result.SpawnersImported)
	fmt.Printf("  Dialogs:     %d\n", result.DialogsImported)
	fmt.Printf("  Rooms:       %d\n", result.RoomsImported)
	fmt.Printf("  Classes:     %d\n", result.ClassesLoaded)
	fmt.Printf("  Assets:      %d\n", result.AssetsImported)
	fmt.Printf("  Characters:  %d relocated (room no longer exists)\n", result.CharactersRelocated)
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
		if err := contenthealth.RecordBaseline(repos, importPath); err != nil {
			log.Fatalf("Failed to record content baseline: %v", err)
		}
		fmt.Println("Import completed successfully!")
	} else {
		fmt.Println("Import completed with errors.")
		os.Exit(1)
	}
}

func runCheck(folderName string, asJSON bool, failOn string) int {
	if failOn != "error" && failOn != "warning" {
		log.Fatalf("invalid -fail-on %q (use error or warning)", failOn)
	}
	importPath := filepath.Join("import", folderName)
	if _, err := os.Stat(importPath); os.IsNotExist(err) {
		log.Fatalf("Import folder not found: %s", importPath)
	}
	if _, err := classkit.LoadDir(filepath.Join(importPath, "data", "classes")); err != nil {
		log.Fatalf("class kit: %v", err)
	}
	converted, err := importer.LoadConverted(importPath)
	if err != nil {
		log.Fatalf("check failed: %v", err)
	}
	rules, err := contenthealth.LoadPackRules(filepath.Join(importPath, "data", "rules"))
	if err != nil {
		log.Fatalf("rules: %v", err)
	}
	// -check does not open the database, so settings are not available.
	// A live health run passes settings.StartRoomID and falls back to this default.
	world := contenthealth.World{
		StartRoomID:   service.DefaultStartRoomID,
		ContentCommit: contenthealth.ContentCommit(importPath),
		Rooms:         converted.Rooms,
		NPCs:          converted.NPCs,
		Items:         converted.Items,
		LootTables:    converted.LootTables,
		Spawners:      converted.Spawners,
		Dialogs:       converted.Dialogs,
		Quests:        converted.Quests,
		Scripts:       converted.Scripts,
		Skills:        converted.Skills,
		Classes:       contenthealth.CatalogClasses(),
	}
	report := contenthealth.Run(world, contenthealth.Options{Rules: rules})
	if asJSON {
		encoded, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			log.Fatalf("json: %v", err)
		}
		fmt.Println(string(encoded))
	} else {
		fmt.Print(contenthealth.FormatText(report, 30))
	}
	if contenthealth.Failed(report, failOn) {
		return 1
	}
	return 0
}
