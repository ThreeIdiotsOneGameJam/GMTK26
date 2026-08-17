package ui

import (
	"reflect"
	"testing"

	"github.com/threeidiotsonegamejam/gmtk26/src/game"
)

func TestTileHoverLines(t *testing.T) {
	tests := []struct {
		name            string
		cell            *game.Cell
		hex             game.Hex
		showCoordinates bool
		want            []string
	}{
		{
			name: "nil cell",
			cell: nil,
		},
		{
			name: "tile without contents",
			cell: &game.Cell{Tile: game.TilePlains, Owner: -1},
			want: []string{
				"Tile: Plains",
				"Territory: Unclaimed",
			},
		},
		{
			name:            "coordinates in debug mode",
			cell:            &game.Cell{Tile: game.TileDesert, Owner: -1},
			hex:             game.NewHex(12, 34),
			showCoordinates: true,
			want: []string{
				"Tile: Desert",
				"Territory: Unclaimed",
				"Coordinates: (12, 34)",
			},
		},
		{
			name: "building",
			cell: &game.Cell{
				Tile:     game.TilePlains,
				Owner:    1,
				Building: &game.BuildingData{Type: game.BuildingTownhall, HP: game.BuildingMaxHP(game.BuildingTownhall)},
			},
			want: []string{
				"Tile: Plains",
				"Territory: Faction 2",
				"",
				"Building: Townhall - Coin x 1",
			},
		},
		{
			name: "resource",
			cell: &game.Cell{Tile: game.TileRock, Owner: -1},
			want: []string{
				"Tile: Rock",
				"Resource: Stone",
				"Territory: Unclaimed",
			},
		},
		{
			name: "unit",
			cell: &game.Cell{
				Tile:  game.TilePlains,
				Owner: -1,
				Units: []game.UnitData{{Type: game.UnitScout, Owner: 2, HP: 3}},
			},
			want: []string{
				"Tile: Plains",
				"Territory: Unclaimed",
				"",
				"Unit: Scout - Faction 3",
			},
		},
		{
			name: "building without output",
			cell: &game.Cell{
				Tile:     game.TilePlains,
				Owner:    0,
				Building: &game.BuildingData{Type: game.BuildingBarracks, HP: game.BuildingMaxHP(game.BuildingBarracks)},
			},
			want: []string{
				"Tile: Plains",
				"Territory: Faction 1",
				"",
				"Building: Barracks",
			},
		},
		{
			name: "farm food output",
			cell: &game.Cell{
				Tile:     game.TilePlains,
				Owner:    0,
				Building: &game.BuildingData{Type: game.BuildingFarm, HP: game.BuildingMaxHP(game.BuildingFarm)},
			},
			want: []string{
				"Tile: Plains",
				"Territory: Faction 1",
				"",
				"Building: Farm - Food x 1",
			},
		},
		{
			name: "building resource and unit",
			cell: &game.Cell{
				Tile:     game.TileGold,
				Owner:    0,
				Building: &game.BuildingData{Type: game.BuildingMine, HP: game.BuildingMaxHP(game.BuildingMine)},
				Units:    []game.UnitData{{Type: game.UnitKnight, Owner: 1, HP: 5}},
			},
			want: []string{
				"Tile: Gold",
				"Resource: Coins + Gold",
				"Territory: Faction 1",
				"",
				"Building: Mine - Coin x 1, Gold x 1",
				"Unit: Knight - Faction 2",
			},
		},
		{
			name: "bank conversion",
			cell: &game.Cell{
				Tile:     game.TilePlains,
				Owner:    0,
				Building: &game.BuildingData{Type: game.BuildingBank, HP: game.BuildingMaxHP(game.BuildingBank)},
			},
			want: []string{
				"Tile: Plains",
				"Territory: Faction 1",
				"",
				"Building: Bank - Coin x 5, Gold -5",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := tileHoverLines(test.cell, test.hex, test.showCoordinates); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("tileHoverLines() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestTileHoverTooltipVisibleOnlyOverWorld(t *testing.T) {
	tests := []struct {
		name               string
		uiBlocksWorldInput bool
		uiModalBlocksInput bool
		want               bool
	}{
		{name: "world", want: true},
		{name: "ui", uiBlocksWorldInput: true},
		{name: "modal", uiModalBlocksInput: true},
		{name: "ui under modal", uiBlocksWorldInput: true, uiModalBlocksInput: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := tileHoverTooltipVisible(
				test.uiBlocksWorldInput,
				test.uiModalBlocksInput,
			)
			if got != test.want {
				t.Fatalf("tileHoverTooltipVisible() = %t, want %t", got, test.want)
			}
		})
	}
}

func TestTileResourceLabel(t *testing.T) {
	tests := map[game.TileType]string{
		game.TileForest: "Wood",
		game.TileJungle: "Wood",
		game.TileRock:   "Stone",
		game.TileCoal:   "",
		game.TileIron:   "Iron",
		game.TileGold:   "Coins + Gold",
		game.TilePlains: "",
		game.TileWater:  "",
	}

	for tile, want := range tests {
		if got := tileResourceLabel(tile); got != want {
			t.Errorf("tileResourceLabel(%s) = %q, want %q", tile, got, want)
		}
	}
}

func TestPlayerFacingResourceOrderOmitsCoalAndSteel(t *testing.T) {
	want := []game.ResourceType{
		game.ResourceFood,
		game.ResourceWood,
		game.ResourceStone,
		game.ResourceIron,
		game.ResourceGold,
	}
	if !reflect.DeepEqual(buildingResourceDisplayOrder, want) {
		t.Fatalf("buildingResourceDisplayOrder = %v, want %v", buildingResourceDisplayOrder, want)
	}
}

func TestUnitCostLabelsUseBalanceData(t *testing.T) {
	tests := map[game.UnitType]string{
		game.UnitScout:   "Scout 8c",
		game.UnitPeasant: "Peasant 8c 4f",
		game.UnitArcher:  "Archer 14c 6f 4w",
		game.UnitKnight:  "Knight 22c 8f 4i",
	}
	for unit, want := range tests {
		if got := unitCostLabel(unit); got != want {
			t.Errorf("unitCostLabel(%s) = %q, want %q", unit, got, want)
		}
	}
}
