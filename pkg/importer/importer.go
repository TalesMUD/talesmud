package importer

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/talesmud/talesmud/pkg/entities/dialogs"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/scripts"
)

// WorldImporter handles importing world data from YAML files
type WorldImporter struct {
	repos      repository.Factory
	importPath string
	verbose    bool
	dryRun     bool
	errors     []string
}

// ImportResult contains the results of an import operation
type ImportResult struct {
	Backup        string
	RoomsImported int
	ItemsImported int
	NPCsImported  int
	ScriptsImported int
	DialogsImported int
	LootTablesImported int
	CharactersRelocated int
	AssetsImported int
	Errors        []string
	Duration      time.Duration
}

// New creates a new WorldImporter
func New(repos repository.Factory, importPath string) *WorldImporter {
	return &WorldImporter{
		repos:      repos,
		importPath: importPath,
		errors:     make([]string, 0),
	}
}

// SetVerbose enables verbose output
func (w *WorldImporter) SetVerbose(v bool) {
	w.verbose = v
}

// SetDryRun enables dry-run mode (validate only)
func (w *WorldImporter) SetDryRun(d bool) {
	w.dryRun = d
}

// Import performs the full import process
func (w *WorldImporter) Import() (*ImportResult, error) {
	start := time.Now()
	result := &ImportResult{}

	// Validate import folder structure
	if err := w.validateImportFolder(); err != nil {
		return nil, fmt.Errorf("invalid import folder: %w", err)
	}

	// Load all data from YAML files
	log.Info("Loading YAML files...")
	yamlScripts, err := w.loadScripts()
	if err != nil {
		w.addError("Failed to load scripts: %v", err)
	}
	yamlItems, err := w.loadItems()
	if err != nil {
		w.addError("Failed to load items: %v", err)
	}
	yamlLootTables, err := w.loadLootTables()
	if err != nil {
		w.addError("Failed to load loot tables: %v", err)
	}
	yamlNPCs, err := w.loadNPCs()
	if err != nil {
		w.addError("Failed to load NPCs: %v", err)
	}
	yamlDialogs, err := w.loadDialogs()
	if err != nil {
		w.addError("Failed to load dialogs: %v", err)
	}
	yamlRooms, err := w.loadRooms()
	if err != nil {
		w.addError("Failed to load rooms: %v", err)
	}

	log.WithFields(log.Fields{
		"scripts":     len(yamlScripts),
		"items":       len(yamlItems),
		"lootTables":  len(yamlLootTables),
		"npcs":        len(yamlNPCs),
		"dialogs":     len(yamlDialogs),
		"rooms":       len(yamlRooms),
	}).Info("Loaded YAML data")

	if w.dryRun {
		log.Info("Dry-run mode: skipping actual import")
		result.Errors = w.errors
		result.Duration = time.Since(start)
		return result, nil
	}

	// Create backup before clearing data
	log.Info("Creating backup...")
	backupPath, err := w.createBackup()
	if err != nil {
		log.WithError(err).Warn("Failed to create backup, continuing anyway")
	} else {
		result.Backup = backupPath
		log.WithField("path", backupPath).Info("Backup created")
	}

	// Clear existing world data (preserve users and characters)
	log.Info("Clearing existing world data...")
	if err := w.clearWorldData(); err != nil {
		return nil, fmt.Errorf("failed to clear world data: %w", err)
	}

	// Import in dependency order
	log.Info("Importing scripts...")
	result.ScriptsImported = w.importScripts(yamlScripts)

	log.Info("Importing items...")
	result.ItemsImported = w.importItems(yamlItems)

	log.Info("Importing loot tables...")
	result.LootTablesImported = w.importLootTables(yamlLootTables)

	log.Info("Importing dialogs...")
	result.DialogsImported = w.importDialogs(yamlDialogs)

	log.Info("Importing NPCs...")
	result.NPCsImported = w.importNPCs(yamlNPCs)

	log.Info("Importing rooms...")
	result.RoomsImported = w.importRooms(yamlRooms)

	// Copy assets
	log.Info("Copying assets...")
	result.AssetsImported, err = w.copyAssets()
	if err != nil {
		w.addError("Failed to copy assets: %v", err)
	}

	// Relocate characters to starting room
	log.Info("Relocating characters to starting room...")
	result.CharactersRelocated, err = w.relocateCharacters()
	if err != nil {
		w.addError("Failed to relocate characters: %v", err)
	}

	result.Errors = w.errors
	result.Duration = time.Since(start)

	return result, nil
}

func (w *WorldImporter) addError(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	w.errors = append(w.errors, msg)
	log.Warn(msg)
}

func (w *WorldImporter) validateImportFolder() error {
	dataPath := filepath.Join(w.importPath, "data")
	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		return fmt.Errorf("data folder not found: %s", dataPath)
	}
	return nil
}

// createBackup creates a JSON backup of current world data
func (w *WorldImporter) createBackup() (string, error) {
	backupDir := "backups"
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", err
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupPath := filepath.Join(backupDir, fmt.Sprintf("world_backup_%s.json", timestamp))

	backup := make(map[string]interface{})

	// Backup rooms
	if rooms, err := w.repos.Rooms().FindAll(); err == nil {
		backup["rooms"] = rooms
	}

	// Backup items
	if items, err := w.repos.Items().FindAll(repository.ItemsQuery{}); err == nil {
		backup["items"] = items
	}

	// Backup NPCs
	if npcs, err := w.repos.NPCs().FindAll(); err == nil {
		backup["npcs"] = npcs
	}

	// Backup scripts
	if scripts, err := w.repos.Scripts().FindAll(); err == nil {
		backup["scripts"] = scripts
	}

	// Backup dialogs
	if dialogs, err := w.repos.Dialogs().FindAll(); err == nil {
		backup["dialogs"] = dialogs
	}

	// Backup loot tables
	if lootTables, err := w.repos.LootTables().FindAll(); err == nil {
		backup["lootTables"] = lootTables
	}

	data, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return "", err
	}

	return backupPath, nil
}

// clearWorldData clears all world data except users and characters
func (w *WorldImporter) clearWorldData() error {
	// Drop in reverse dependency order
	if err := w.repos.NPCSpawners().Drop(); err != nil {
		log.WithError(err).Warn("Failed to drop NPC spawners")
	}
	if err := w.repos.Rooms().Drop(); err != nil {
		return fmt.Errorf("failed to drop rooms: %w", err)
	}
	if err := w.repos.NPCs().Drop(); err != nil {
		return fmt.Errorf("failed to drop NPCs: %w", err)
	}
	if err := w.repos.Dialogs().Drop(); err != nil {
		return fmt.Errorf("failed to drop dialogs: %w", err)
	}
	if err := w.repos.LootTables().Drop(); err != nil {
		return fmt.Errorf("failed to drop loot tables: %w", err)
	}
	if err := w.repos.Items().Drop(); err != nil {
		return fmt.Errorf("failed to drop items: %w", err)
	}
	if err := w.repos.Scripts().Drop(); err != nil {
		return fmt.Errorf("failed to drop scripts: %w", err)
	}

	return nil
}

// YAML loading functions

func (w *WorldImporter) loadYAMLFiles(subdir string, loader func([]byte) error) error {
	dir := filepath.Join(w.importPath, "data", subdir)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil // Directory doesn't exist, skip
	}

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(path), ".yaml") && !strings.HasSuffix(strings.ToLower(path), ".yml") {
			return nil
		}
		// Skip Zone.Identifier files
		if strings.Contains(path, "Zone.Identifier") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			w.addError("Failed to read %s: %v", path, err)
			return nil
		}

		if err := loader(data); err != nil {
			w.addError("Failed to parse %s: %v", path, err)
		}

		return nil
	})
}

func (w *WorldImporter) loadScripts() ([]*YAMLScript, error) {
	var result []*YAMLScript
	err := w.loadYAMLFiles("scripts", func(data []byte) error {
		var s YAMLScript
		if err := yaml.Unmarshal(data, &s); err != nil {
			return err
		}
		result = append(result, &s)
		return nil
	})
	return result, err
}

func (w *WorldImporter) loadItems() ([]*YAMLItem, error) {
	var result []*YAMLItem
	err := w.loadYAMLFiles("items", func(data []byte) error {
		var item YAMLItem
		if err := yaml.Unmarshal(data, &item); err != nil {
			return err
		}
		result = append(result, &item)
		return nil
	})
	return result, err
}

func (w *WorldImporter) loadLootTables() ([]*YAMLLootTable, error) {
	var result []*YAMLLootTable
	seen := make(map[string]bool)
	err := w.loadYAMLFiles("loot_tables", func(data []byte) error {
		var lt YAMLLootTable
		if err := yaml.Unmarshal(data, &lt); err != nil {
			return err
		}
		if !seen[lt.ID] {
			seen[lt.ID] = true
			result = append(result, &lt)
		}
		return nil
	})
	return result, err
}

func (w *WorldImporter) loadNPCs() ([]*YAMLNPC, error) {
	var result []*YAMLNPC
	err := w.loadYAMLFiles("npcs", func(data []byte) error {
		var n YAMLNPC
		if err := yaml.Unmarshal(data, &n); err != nil {
			return err
		}
		result = append(result, &n)
		return nil
	})
	return result, err
}

func (w *WorldImporter) loadDialogs() ([]*YAMLDialog, error) {
	var result []*YAMLDialog
	seen := make(map[string]bool)
	err := w.loadYAMLFiles("dialogs", func(data []byte) error {
		var d YAMLDialog
		if err := yaml.Unmarshal(data, &d); err != nil {
			return err
		}
		if !seen[d.ID] {
			seen[d.ID] = true
			result = append(result, &d)
		}
		return nil
	})
	return result, err
}

func (w *WorldImporter) loadRooms() ([]*YAMLRoom, error) {
	var result []*YAMLRoom
	err := w.loadYAMLFiles("rooms", func(data []byte) error {
		var r YAMLRoom
		if err := yaml.Unmarshal(data, &r); err != nil {
			return err
		}
		result = append(result, &r)
		return nil
	})
	return result, err
}

// Import functions

func (w *WorldImporter) importScripts(yamlScripts []*YAMLScript) int {
	count := 0
	for _, s := range yamlScripts {
		entity := s.ToEntity()
		if _, err := w.repos.Scripts().Import(entity); err != nil {
			w.addError("Failed to import script %s: %v", s.ID, err)
		} else {
			count++
			if w.verbose {
				log.WithField("id", s.ID).Debug("Imported script")
			}
		}
	}
	return count
}

func (w *WorldImporter) importItems(yamlItems []*YAMLItem) int {
	count := 0
	for _, item := range yamlItems {
		entity := item.ToEntity()
		if _, err := w.repos.Items().Import(entity); err != nil {
			w.addError("Failed to import item %s: %v", item.ID, err)
		} else {
			count++
			if w.verbose {
				log.WithField("id", item.ID).Debug("Imported item")
			}
		}
	}
	return count
}

func (w *WorldImporter) importLootTables(yamlLootTables []*YAMLLootTable) int {
	count := 0
	for _, lt := range yamlLootTables {
		entity := lt.ToEntity()
		if _, err := w.repos.LootTables().Import(entity); err != nil {
			w.addError("Failed to import loot table %s: %v", lt.ID, err)
		} else {
			count++
			if w.verbose {
				log.WithField("id", lt.ID).Debug("Imported loot table")
			}
		}
	}
	return count
}

func (w *WorldImporter) importDialogs(yamlDialogs []*YAMLDialog) int {
	count := 0
	for _, d := range yamlDialogs {
		entity := d.ToEntity()
		if _, err := w.repos.Dialogs().Import(entity); err != nil {
			w.addError("Failed to import dialog %s: %v", d.ID, err)
		} else {
			count++
			if w.verbose {
				log.WithField("id", d.ID).Debug("Imported dialog")
			}
		}
	}
	return count
}

func (w *WorldImporter) importNPCs(yamlNPCs []*YAMLNPC) int {
	count := 0
	for _, n := range yamlNPCs {
		entity := n.ToEntity()
		if _, err := w.repos.NPCs().Import(entity); err != nil {
			w.addError("Failed to import NPC %s: %v", n.ID, err)
		} else {
			count++
			if w.verbose {
				log.WithField("id", n.ID).Debug("Imported NPC")
			}
		}
	}
	return count
}

func (w *WorldImporter) importRooms(yamlRooms []*YAMLRoom) int {
	count := 0
	for _, r := range yamlRooms {
		entity := r.ToEntity()
		if _, err := w.repos.Rooms().Import(entity); err != nil {
			w.addError("Failed to import room %s: %v", r.ID, err)
		} else {
			count++
			if w.verbose {
				log.WithField("id", r.ID).Debug("Imported room")
			}
		}
	}
	return count
}

// copyAssets copies room images to the backgrounds folder
func (w *WorldImporter) copyAssets() (int, error) {
	srcDir := filepath.Join(w.importPath, "assets", "images", "rooms")
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return 0, nil // No assets to copy
	}

	dstDir := os.Getenv("BACKGROUNDS_PATH")
	if dstDir == "" {
		dstDir = "./uploads/backgrounds"
	}

	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return 0, fmt.Errorf("failed to create backgrounds directory: %w", err)
	}

	count := 0
	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		// Only copy image files
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".webp" {
			return nil
		}
		// Skip Zone.Identifier files
		if strings.Contains(path, "Zone.Identifier") {
			return nil
		}

		dstPath := filepath.Join(dstDir, info.Name())
		if err := copyFile(path, dstPath); err != nil {
			w.addError("Failed to copy asset %s: %v", path, err)
			return nil
		}
		count++
		if w.verbose {
			log.WithField("file", info.Name()).Debug("Copied asset")
		}
		return nil
	})

	return count, err
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// relocateCharacters moves all characters to the starting room (R0001)
func (w *WorldImporter) relocateCharacters() (int, error) {
	startRoomID := "R0001"

	chars, err := w.repos.Characters().FindAll()
	if err != nil {
		return 0, fmt.Errorf("failed to find characters: %w", err)
	}

	count := 0
	for _, char := range chars {
		char.CurrentRoomID = startRoomID
		char.BoundRoomID = startRoomID
		if err := w.repos.Characters().Update(char.ID, char); err != nil {
			w.addError("Failed to relocate character %s: %v", char.Name, err)
		} else {
			count++
			if w.verbose {
				log.WithField("name", char.Name).Debug("Relocated character")
			}
		}
	}

	return count, nil
}

// Exported entity types for use in main
type (
	Script    = scripts.Script
	Item      = items.Item
	LootTable = items.LootTable
	NPC       = npc.NPC
	Dialog    = dialogs.Dialog
	Room      = rooms.Room
)
