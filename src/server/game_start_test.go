package server

import (
	"reflect"
	"testing"

	"github.com/threeidiotsonegamejam/gmtk26/src/game"
)

func TestAssignedStartsAreDeterministicAndWellSpaced(t *testing.T) {
	first := generatedStartGame(8675309)
	second := generatedStartGame(8675309)

	firstStarts := townhallStarts(first)
	secondStarts := townhallStarts(second)
	if len(firstStarts) != len(first.Factions) {
		t.Fatalf("Townhall starts = %d, want %d", len(firstStarts), len(first.Factions))
	}
	if !reflect.DeepEqual(firstStarts, secondStarts) {
		t.Fatalf("same seed produced different starts: %v vs %v", firstStarts, secondStarts)
	}

	island := first.Map.LargestLandIsland()
	for faction, start := range firstStarts {
		cell := first.Map.GetCell(start)
		if cell.Tile != game.TilePlains {
			t.Errorf("faction %d started on %s, want Plains", faction, cell.Tile)
		}
		if !island[start] {
			t.Errorf("faction %d started outside largest island at %v", faction, start)
		}
		if quality := startingZoneQuality(&first.Map, start); quality < 4 {
			t.Errorf("faction %d starting-zone quality = %d, want at least 4", faction, quality)
		}
	}
	for i := 0; i < len(firstStarts); i++ {
		for j := i + 1; j < len(firstStarts); j++ {
			if distance := firstStarts[i].Distance(firstStarts[j]); distance < startZoneMinimumSpacing {
				t.Errorf("start distance %d-%d = %d, want at least %d", i, j, distance, startZoneMinimumSpacing)
			}
		}
	}
}

func generatedStartGame(seed int64) *game.Game {
	g := &game.Game{Map: game.Map{Seed: seed}}
	g.Map.Generate()
	gi := NewGameInstance(1, g, make([]*Client, len(g.Factions)))
	gi.assignStartingCells()
	return g
}

func townhallStarts(g *game.Game) []game.Hex {
	starts := make([]game.Hex, len(g.Factions))
	found := make([]bool, len(g.Factions))
	for x := range g.Map.Grid {
		for y := range g.Map.Grid[x] {
			cell := &g.Map.Grid[x][y]
			if cell.BuildingType() != game.BuildingTownhall ||
				cell.Owner < 0 ||
				int(cell.Owner) >= len(starts) {
				continue
			}
			starts[cell.Owner] = game.NewHex(int32(x), int32(y))
			found[cell.Owner] = true
		}
	}
	for _, ok := range found {
		if !ok {
			return nil
		}
	}
	return starts
}
