package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/exporter"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/scripts/runner"
	"github.com/talesmud/talesmud/pkg/service"
)

func main() {
	inputPath := flag.String("input", "", "Path to export JSON file")
	sqlitePath := flag.String("sqlite", "talesmud.db", "SQLite database path")
	dropFirst := flag.Bool("drop", true, "Drop existing data before import")
	flag.Parse()

	if *inputPath == "" {
		log.Fatal("missing -input")
	}

	file, err := os.Open(*inputPath)
	if err != nil {
		log.WithError(err).Fatal("failed to open input file")
	}
	defer file.Close()

	var data exporter.Data
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		log.WithError(err).Fatal("failed to decode export JSON")
	}

	client, err := dbsqlite.Open(*sqlitePath)
	if err != nil {
		log.WithError(err).Fatal("failed to open sqlite database")
	}
	defer client.Close()

	repos := repository.NewSQLiteFactory(client)
	scriptRunner := runner.NewDefaultScriptRunner()
	facade := service.NewFacade(repos, scriptRunner)
	scriptRunner.SetServices(facade, nil)

	if *dropFirst {
		_ = facade.RoomsService().Drop()
		_ = facade.CharactersService().Drop()
		_ = facade.UsersService().Drop()
		_ = facade.ItemsService().Drop()
		_ = facade.ScriptsService().Drop()
		_ = facade.NPCsService().Drop()
		_ = facade.DialogsService().Drop()
		_ = facade.LootTablesService().Drop()
	}

	for _, room := range data.Rooms {
		_, _ = facade.RoomsService().Import(room)
	}
	for _, character := range data.Characters {
		_, _ = facade.CharactersService().Import(character)
	}
	for _, user := range data.Users {
		_, _ = facade.UsersService().Import(user)
	}
	for _, item := range data.Items {
		_, _ = facade.ItemsService().Import(item)
	}
	for _, script := range data.Scripts {
		_, _ = facade.ScriptsService().Import(script)
	}
	for _, npc := range data.NPCs {
		_, _ = facade.NPCsService().Import(npc)
	}
	for _, dialog := range data.Dialogs {
		_, _ = facade.DialogsService().Import(dialog)
	}
	for _, party := range data.Parties {
		_, _ = facade.PartiesService().Store(party)
	}
	for _, lootTable := range data.LootTables {
		_, _ = facade.LootTablesService().Import(lootTable)
	}

	fmt.Println("Migration completed successfully.")
}
