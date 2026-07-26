package game

import (
	"testing"

	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

func movementTestMap(width, height int32) Map {
	grid := make([][]Cell, width)
	for x := range grid {
		grid[x] = make([]Cell, height)
		for y := range grid[x] {
			grid[x][y] = Cell{Tile: TilePlains, Owner: -1}
		}
	}
	return Map{
		Grid:     grid,
		GridSize: vec.Vec2i{X: width, Y: height},
	}
}

func TestUnitMovementBudgets(t *testing.T) {
	tests := map[UnitType]int{
		UnitScout:   3,
		UnitPeasant: 2,
		UnitArcher:  2,
		UnitKnight:  4,
		UnitUnknown: 0,
	}
	for unit, want := range tests {
		if got := UnitMovementBudget(unit); got != want {
			t.Fatalf("%s budget = %d, want %d", unit, got, want)
		}
	}
}

func TestTerrainMovementCosts(t *testing.T) {
	tests := map[TileType]int{
		TilePlains: 1,
		TileRock:   1,
		TileIron:   1,
		TileCoal:   1,
		TileGold:   1,
		TileForest: 2,
		TileDesert: 3,
		TileJungle: 3,
		TileWater:  0,
		TileVoid:   0,
	}
	for tile, want := range tests {
		if got := TerrainMovementCost(tile); got != want {
			t.Fatalf("%s cost = %d, want %d", tile, got, want)
		}
	}
}

func TestAdvanceUnitPathGuaranteesFirstCostlyStep(t *testing.T) {
	m := movementTestMap(1, 4)
	m.Grid[0][1].Tile = TileJungle
	path := []Hex{NewHex(0, 0), NewHex(0, 1), NewHex(0, 2)}

	got := m.AdvanceUnitPath(path, 2)
	if len(got) != 2 || got[1] != NewHex(0, 1) {
		t.Fatalf("traversed = %v, want costly first step only", got)
	}
}

func TestAdvanceUnitPathRespectsRemainingBudget(t *testing.T) {
	m := movementTestMap(1, 4)
	m.Grid[0][2].Tile = TileForest
	path := []Hex{
		NewHex(0, 0),
		NewHex(0, 1),
		NewHex(0, 2),
		NewHex(0, 3),
	}

	got := m.AdvanceUnitPath(path, 2)
	if len(got) != 2 || got[1] != NewHex(0, 1) {
		t.Fatalf("traversed = %v, want stop before forest", got)
	}
}

func TestFindUnitPathUsesTerrainOwnershipAndBlockers(t *testing.T) {
	m := movementTestMap(1, 4)
	start := NewHex(0, 0)
	goal := NewHex(0, 3)
	m.GetCell(start).Units = []UnitData{{Type: UnitScout, Owner: 0, HP: 3}}

	if _, ok := m.FindUnitPath(0, start, goal); !ok {
		t.Fatal("expected clear friendly/unclaimed path")
	}

	m.GetCell(NewHex(0, 1)).Owner = 1
	if _, ok := m.FindUnitPath(0, start, goal); ok {
		t.Fatal("path crossed enemy-owned tile")
	}

	m.GetCell(NewHex(0, 1)).Owner = -1
	m.GetCell(NewHex(0, 1)).Units = []UnitData{{Type: UnitPeasant, Owner: 0, HP: 3}}
	if _, ok := m.FindUnitPath(0, start, goal); ok {
		t.Fatal("path crossed friendly unit blocker")
	}
}

func TestFindUnitPathChoosesCheaperTerrain(t *testing.T) {
	m := movementTestMap(3, 2)
	start := NewHex(0, 1)
	goal := NewHex(2, 1)
	m.GetCell(start).Units = []UnitData{{Type: UnitKnight, Owner: 0, HP: 5}}
	m.GetCell(NewHex(1, 1)).Tile = TileJungle

	path, ok := m.FindUnitPath(0, start, goal)
	if !ok {
		t.Fatal("expected route")
	}
	for _, hex := range path {
		if hex == NewHex(1, 1) {
			t.Fatalf("path used expensive direct tile: %v", path)
		}
	}
}

func TestMovementTurnStops(t *testing.T) {
	m := movementTestMap(1, 6)
	path := []Hex{
		NewHex(0, 0),
		NewHex(0, 1),
		NewHex(0, 2),
		NewHex(0, 3),
		NewHex(0, 4),
		NewHex(0, 5),
	}
	got := m.MovementTurnStops(path, 4)
	want := []Hex{NewHex(0, 4), NewHex(0, 5)}
	if len(got) != len(want) {
		t.Fatalf("stops = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("stops = %v, want %v", got, want)
		}
	}
}
