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
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/skills"
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
	Backup              string
	RoomsImported       int
	ItemsImported       int
	NPCsImported        int
	SpawnersImported    int
	ScriptsImported     int
	DialogsImported     int
	LootTablesImported  int
	QuestsImported      int
	SkillsImported      int
	CharactersRelocated int
	AssetsImported      int
	ValidationWarnings  int
	Errors              []string
	Duration            time.Duration
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
	yamlSpawners, err := w.loadSpawners()
	if err != nil {
		w.addError("Failed to load spawners: %v", err)
	}
	yamlQuests, err := w.loadQuests()
	if err != nil {
		w.addError("Failed to load quests: %v", err)
	}
	yamlSkills, err := w.loadSkills()
	if err != nil {
		w.addError("Failed to load skills: %v", err)
	}

	log.WithFields(log.Fields{
		"scripts":    len(yamlScripts),
		"items":      len(yamlItems),
		"lootTables": len(yamlLootTables),
		"npcs":       len(yamlNPCs),
		"dialogs":    len(yamlDialogs),
		"rooms":      len(yamlRooms),
		"spawners":   len(yamlSpawners),
		"quests":     len(yamlQuests),
		"skills":     len(yamlSkills),
	}).Info("Loaded YAML data")

	// Validate cross-references and script code
	log.Info("Validating data consistency...")
	validationWarnings := w.validateData(
		yamlRooms, yamlItems, yamlNPCs, yamlScripts,
		yamlDialogs, yamlLootTables, yamlSpawners, yamlQuests, yamlSkills,
	)
	result.ValidationWarnings = validationWarnings
	if validationWarnings > 0 {
		log.WithField("warnings", validationWarnings).Warn("Data validation found issues")
	} else {
		log.Info("Data validation passed — no issues found")
	}

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
	result.ItemsImported = w.importItems(yamlItems, roomPlacedItemIDs(yamlRooms))

	log.Info("Importing loot tables...")
	result.LootTablesImported = w.importLootTables(yamlLootTables)

	log.Info("Importing dialogs...")
	result.DialogsImported = w.importDialogs(yamlDialogs)

	log.Info("Importing NPCs...")
	result.NPCsImported = w.importNPCs(yamlNPCs)

	log.Info("Importing rooms...")
	result.RoomsImported = w.importRooms(yamlRooms)

	log.Info("Importing NPC spawners...")
	result.SpawnersImported = w.importSpawners(yamlSpawners)

	log.Info("Importing quests...")
	result.QuestsImported = w.importQuests(yamlQuests)

	log.Info("Importing skills...")
	result.SkillsImported = w.importSkills(yamlSkills)

	// Copy assets
	log.Info("Copying assets...")
	result.AssetsImported, err = w.copyAssets()
	if err != nil {
		w.addError("Failed to copy assets: %v", err)
	}

	// Persist start room so guests spawn in R0001, not rooms[0]
	if err := w.applyStartRoom(yamlRooms); err != nil {
		w.addError("Failed to set start room: %v", err)
	}

	// Relocate characters whose rooms no longer exist
	log.Info("Checking character room assignments...")
	result.CharactersRelocated, err = w.relocateCharacters(yamlRooms)
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

	// Backup NPC spawners
	if spawners, err := w.repos.NPCSpawners().FindAll(); err == nil {
		backup["npcSpawners"] = spawners
	}

	// Backup quests
	if quests, err := w.repos.Quests().FindAll(); err == nil {
		backup["quests"] = quests
	}

	// Backup skills
	if skills, err := w.repos.Skills().FindAll(); err == nil {
		backup["skills"] = skills
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
	if err := w.repos.Quests().Drop(); err != nil {
		log.WithError(err).Warn("Failed to drop quests")
	}
	if err := w.repos.Skills().Drop(); err != nil {
		log.WithError(err).Warn("Failed to drop skills")
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

func (w *WorldImporter) loadSpawners() ([]*YAMLSpawner, error) {
	var result []*YAMLSpawner
	err := w.loadYAMLFiles("npc_spawners", func(data []byte) error {
		var s YAMLSpawner
		if err := yaml.Unmarshal(data, &s); err != nil {
			return err
		}
		result = append(result, &s)
		return nil
	})
	return result, err
}

func (w *WorldImporter) loadQuests() ([]*YAMLQuest, error) {
	var result []*YAMLQuest
	err := w.loadYAMLFiles("quests", func(data []byte) error {
		var q YAMLQuest
		if err := yaml.Unmarshal(data, &q); err != nil {
			return err
		}
		result = append(result, &q)
		return nil
	})
	return result, err
}

func (w *WorldImporter) loadSkills() ([]*YAMLSkill, error) {
	var result []*YAMLSkill
	err := w.loadYAMLFiles("skills", func(data []byte) error {
		var s YAMLSkill
		if err := yaml.Unmarshal(data, &s); err != nil {
			return err
		}
		result = append(result, &s)
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

func roomPlacedItemIDs(yamlRooms []*YAMLRoom) map[string]bool {
	ids := make(map[string]bool)
	for _, r := range yamlRooms {
		for _, item := range r.Items {
			if item.ID != "" {
				ids[item.ID] = true
			}
		}
	}
	return ids
}

func (w *WorldImporter) importItems(yamlItems []*YAMLItem, roomPlaced map[string]bool) int {
	count := 0
	for _, item := range yamlItems {
		entity := item.ToEntity()
		if roomPlaced[item.ID] {
			entity.CopyOnPickup = true
		}
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

func (w *WorldImporter) importSpawners(yamlSpawners []*YAMLSpawner) int {
	count := 0
	for _, s := range yamlSpawners {
		entity := s.ToEntity()
		if _, err := w.repos.NPCSpawners().Import(entity); err != nil {
			w.addError("Failed to import spawner %s: %v", s.ID, err)
		} else {
			count++
			if w.verbose {
				log.WithField("id", s.ID).Debug("Imported spawner")
			}
		}
	}
	return count
}

func (w *WorldImporter) importQuests(yamlQuests []*YAMLQuest) int {
	count := 0
	for _, q := range yamlQuests {
		entity := q.ToEntity()
		if _, err := w.repos.Quests().Import(entity); err != nil {
			w.addError("Failed to import quest %s: %v", q.ID, err)
		} else {
			count++
			if w.verbose {
				log.WithField("id", q.ID).Debug("Imported quest")
			}
		}
	}
	return count
}

func (w *WorldImporter) importSkills(yamlSkills []*YAMLSkill) int {
	count := 0
	for _, s := range yamlSkills {
		entity := s.ToEntity()
		if _, err := w.repos.Skills().Import(entity); err != nil {
			w.addError("Failed to import skill %s: %v", s.ID, err)
		} else {
			count++
			if w.verbose {
				log.WithField("id", s.ID).Debug("Imported skill")
			}
		}
	}
	// Refresh the in-memory cache after import
	if count > 0 {
		if all, err := w.repos.Skills().FindAll(); err == nil {
			skills.RefreshCache(all)
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

// copyAssets copies room images to backgrounds and NPC portraits to portraits.
func (w *WorldImporter) copyAssets() (int, error) {
	bg := os.Getenv("BACKGROUNDS_PATH")
	if bg == "" {
		bg = "./uploads/backgrounds"
	}
	pt := os.Getenv("PORTRAITS_PATH")
	if pt == "" {
		pt = "./uploads/portraits"
	}
	n1, err := w.copyImageDir(filepath.Join(w.importPath, "assets", "images", "rooms"), bg)
	if err != nil {
		return n1, err
	}
	portraitSrcs := []string{
		filepath.Join(w.importPath, "assets", "images", "npcs"),
		filepath.Join(w.importPath, "assets", "images", "sprites", "npcs"),
		filepath.Join(w.importPath, "assets", "images", "sprites", "enemies"),
	}
	n2 := 0
	for _, src := range portraitSrcs {
		c, err := w.copyImageDir(src, pt)
		n2 += c
		if err != nil {
			return n1 + n2, err
		}
	}
	it := os.Getenv("ITEM_ART_PATH")
	if it == "" {
		it = "./uploads/items"
	}
	n3, err := w.copyImageDir(filepath.Join(w.importPath, "assets", "images", "items"), it)
	if err != nil {
		return n1 + n2 + n3, err
	}
	return n1 + n2 + n3, nil
}

func skipAssetPath(path string) bool {
	if strings.Contains(path, "Zone.Identifier") {
		return true
	}
	parts := strings.Split(filepath.ToSlash(path), "/")
	for _, p := range parts {
		if p == "_raw" || p == "previews" {
			return true
		}
	}
	base := strings.ToLower(filepath.Base(path))
	if base == "contact_sheet.png" {
		return true
	}
	if strings.HasSuffix(base, ".tmp") || strings.HasSuffix(base, ".temp") || strings.HasSuffix(base, ".bak") || strings.HasSuffix(base, "~") {
		return true
	}
	return false
}

func (w *WorldImporter) copyImageDir(srcDir, dstDir string) (int, error) {
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return 0, nil
	}
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return 0, fmt.Errorf("failed to create %s: %w", dstDir, err)
	}
	count := 0
	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".webp" {
			return nil
		}
		if skipAssetPath(path) {
			return nil
		}
		if err := copyFile(path, filepath.Join(dstDir, info.Name())); err != nil {
			w.addError("Failed to copy asset %s: %v", path, err)
			return nil
		}
		count++
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

// relocateCharacters checks each character's room against the imported world.
// Characters whose current room still exists are left in place.
// Characters whose room no longer exists are moved to the starting room (R0001).
func (w *WorldImporter) applyStartRoom(yamlRooms []*YAMLRoom) error {
	startRoomID := "R0001"
	found := false
	for _, r := range yamlRooms {
		if r.ID == startRoomID {
			found = true
			break
		}
	}
	if !found {
		return nil
	}

	s, err := w.repos.ServerSettings().Get()
	if err != nil || s == nil {
		return err
	}
	s.StartRoomID = startRoomID
	log.WithField("startRoomID", startRoomID).Info("Setting server start room")
	return w.repos.ServerSettings().Upsert(s)
}

func (w *WorldImporter) relocateCharacters(yamlRooms []*YAMLRoom) (int, error) {
	startRoomID := "R0001"

	// Build a set of valid room IDs from the newly imported rooms
	validRooms := make(map[string]bool, len(yamlRooms))
	for _, r := range yamlRooms {
		validRooms[r.ID] = true
	}

	chars, err := w.repos.Characters().FindAll()
	if err != nil {
		return 0, fmt.Errorf("failed to find characters: %w", err)
	}

	count := 0
	for _, char := range chars {
		needsUpdate := false

		if !validRooms[char.CurrentRoomID] {
			log.WithFields(log.Fields{
				"name":    char.Name,
				"oldRoom": char.CurrentRoomID,
			}).Info("Room no longer exists, relocating to starting room")
			char.CurrentRoomID = startRoomID
			needsUpdate = true
		}

		if !validRooms[char.BoundRoomID] {
			char.BoundRoomID = startRoomID
			needsUpdate = true
		}

		if !needsUpdate {
			if w.verbose {
				log.WithFields(log.Fields{
					"name": char.Name,
					"room": char.CurrentRoomID,
				}).Debug("Character room still exists, keeping in place")
			}
			continue
		}

		if err := w.repos.Characters().Update(char.ID, char); err != nil {
			w.addError("Failed to relocate character %s: %v", char.Name, err)
		} else {
			count++
		}
	}

	return count, nil
}

// Exported entity types for use in main
type (
	Script        = scripts.Script
	Item          = items.Item
	LootTable     = items.LootTable
	NPC           = npc.NPC
	NPCSpawner    = npc.NPCSpawner
	Dialog        = dialogs.Dialog
	Room          = rooms.Room
	Quest         = quests.Quest
	Skill         = skills.Skill
	Objective     = quests.Objective
	ObjectiveType = quests.ObjectiveType
	Reward        = quests.Reward
	QuestSource   = quests.QuestSource
)
