package game

import (
	"testing"

	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

func TestDefaultMapSizeAndCoalFreeGeneration(t *testing.T) {
	for seed := int64(0); seed < 8; seed++ {
		m := Map{Seed: seed}
		m.Generate()
		if m.GridSize != (vec.Vec2i{X: 72, Y: 72}) {
			t.Fatalf("seed %d GridSize = %v, want 72x72", seed, m.GridSize)
		}
		for x := range m.Grid {
			for y := range m.Grid[x] {
				if m.Grid[x][y].Tile == TileCoal {
					t.Fatalf("seed %d generated Coal at (%d,%d)", seed, x, y)
				}
			}
		}
	}
}

func TestSpreadResourcesTransfersCoalShareToIron(t *testing.T) {
	grid := make([][]Cell, 10)
	for x := range grid {
		grid[x] = make([]Cell, 10)
		for y := range grid[x] {
			grid[x][y] = Cell{Tile: TileRock, Owner: -1}
		}
	}
	m := Map{
		Grid:     grid,
		GridSize: vec.Vec2i{X: 10, Y: 10},
	}

	spreadResources(&m, 42)

	counts := make(map[TileType]int)
	for x := range m.Grid {
		for y := range m.Grid[x] {
			counts[m.Grid[x][y].Tile]++
		}
	}
	if counts[TileIron] != 20 {
		t.Fatalf("Iron count = %d, want 20", counts[TileIron])
	}
	if counts[TileGold] != 5 {
		t.Fatalf("Gold count = %d, want 5", counts[TileGold])
	}
	if counts[TileCoal] != 0 {
		t.Fatalf("Coal count = %d, want 0", counts[TileCoal])
	}
}
