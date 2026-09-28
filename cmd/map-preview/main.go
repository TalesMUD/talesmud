// map-preview compiles a read-only room JSON snapshot into a fully explored
// review atlas and terrain census. It never opens a database or starts the game.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/worldmap"
)

func main() {
	input := flag.String("rooms", "", "JSON array of room records (required)")
	current := flag.String("current", "R0201", "current room for mockup")
	fog := flag.String("fog", "", "comma-separated unexplored room IDs")
	flag.Parse()
	if *input == "" {
		fmt.Fprintln(os.Stderr, "-rooms is required")
		os.Exit(2)
	}
	b, err := os.ReadFile(*input)
	if err != nil {
		panic(err)
	}
	var rs []*rooms.Room
	if err := json.Unmarshal(b, &rs); err != nil {
		panic(err)
	}
	ch := &characters.Character{Entity: &entities.Entity{ID: "terrain-preview"}, DiscoveredRooms: map[string]bool{}}
	ch.CurrentRoomID = *current
	counts := map[string]int{}
	fallback := 0
	fallbackIDs := []string{}
	for _, r := range rs {
		t, def := worldmap.ClassifyTerrain(r)
		counts[t]++
		if def {
			fallback++
			fallbackIDs = append(fallbackIDs, r.ID)
		}
		ch.DiscoveredRooms[r.ID] = true
	}
	for _, id := range strings.Split(*fog, ",") {
		delete(ch.DiscoveredRooms, id)
	}
	result := struct {
		Atlas           worldmap.PlayerMap `json:"atlas"`
		Counts          map[string]int     `json:"counts"`
		Total           int                `json:"total"`
		DefaultFallback int                `json:"defaultFallback"`
		FallbackIDs     []string           `json:"fallbackIds"`
	}{worldmap.Reveal(worldmap.Compile(rs), ch), counts, len(rs), fallback, fallbackIDs}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(result); err != nil {
		panic(err)
	}
}
