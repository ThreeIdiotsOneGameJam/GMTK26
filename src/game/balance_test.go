package game

import (
	"reflect"
	"testing"
	"time"
)

func TestMatchRules(t *testing.T) {
	if StartingCoins != 25 {
		t.Fatalf("StartingCoins = %d, want 25", StartingCoins)
	}
	if MatchDuration != 15*time.Minute {
		t.Fatalf("MatchDuration = %s, want 15m", MatchDuration)
	}
	if RoundDuration != 5*time.Second {
		t.Fatalf("RoundDuration = %s, want 5s", RoundDuration)
	}
	if ScoreIntervalRounds != 6 {
		t.Fatalf("ScoreIntervalRounds = %d, want 6", ScoreIntervalRounds)
	}
}

func TestUnitBalance(t *testing.T) {
	tests := []struct {
		unit      UnitType
		coins     int32
		resources Resources
		stats     UnitStats
		movement  int
		score     int32
	}{
		{UnitScout, 8, Resources{}, UnitStats{MaxHP: 3, Attack: 0}, 4, 3},
		{UnitPeasant, 8, Resources{ResourceFood: 4}, UnitStats{MaxHP: 5, Attack: 1}, 2, 5},
		{UnitArcher, 14, Resources{ResourceFood: 6, ResourceWood: 4}, UnitStats{MaxHP: 4, Attack: 2}, 2, 8},
		{UnitKnight, 22, Resources{ResourceFood: 8, ResourceIron: 4}, UnitStats{MaxHP: 8, Attack: 3}, 3, 12},
	}

	for _, test := range tests {
		t.Run(test.unit.String(), func(t *testing.T) {
			if got := UnitCost(test.unit); got != test.coins {
				t.Errorf("UnitCost() = %d, want %d", got, test.coins)
			}
			if got := UnitResourceCost(test.unit); !reflect.DeepEqual(got, test.resources) {
				t.Errorf("UnitResourceCost() = %v, want %v", got, test.resources)
			}
			if got := GetUnitStats(test.unit); got != test.stats {
				t.Errorf("GetUnitStats() = %+v, want %+v", got, test.stats)
			}
			if got := UnitMovementBudget(test.unit); got != test.movement {
				t.Errorf("UnitMovementBudget() = %d, want %d", got, test.movement)
			}
			if got := UnitDestructionScore(test.unit); got != test.score {
				t.Errorf("UnitDestructionScore() = %d, want %d", got, test.score)
			}
		})
	}
}

func TestBuildingBalance(t *testing.T) {
	tests := []struct {
		building  BuildingType
		coins     int32
		resources Resources
		hp        int8
		control   int32
		destroy   int32
	}{
		{BuildingTownhall, 0, Resources{}, 24, 0, 30},
		{BuildingFarm, 12, Resources{}, 10, 2, 8},
		{BuildingForester, 10, Resources{}, 10, 2, 8},
		{BuildingMine, 14, Resources{}, 12, 3, 10},
		{BuildingBarracks, 24, Resources{ResourceWood: 6, ResourceStone: 4}, 16, 3, 15},
		{BuildingBank, 25, Resources{}, 10, 3, 12},
	}

	for _, test := range tests {
		t.Run(test.building.String(), func(t *testing.T) {
			if got := BuildingCost(test.building); got != test.coins {
				t.Errorf("BuildingCost() = %d, want %d", got, test.coins)
			}
			if got := BuildingResourceCost(test.building); !reflect.DeepEqual(got, test.resources) {
				t.Errorf("BuildingResourceCost() = %v, want %v", got, test.resources)
			}
			if got := BuildingMaxHP(test.building); got != test.hp {
				t.Errorf("BuildingMaxHP() = %d, want %d", got, test.hp)
			}
			if got := BuildingControlScore(test.building, TileRock); got != test.control {
				t.Errorf("BuildingControlScore() = %d, want %d", got, test.control)
			}
			if got := BuildingDestructionScore(test.building, TileRock); got != test.destroy {
				t.Errorf("BuildingDestructionScore() = %d, want %d", got, test.destroy)
			}
		})
	}

	if got := BuildingControlScore(BuildingMine, TileGold); got != 5 {
		t.Fatalf("Gold Mine control score = %d, want 5", got)
	}
	if got := BuildingDestructionScore(BuildingMine, TileGold); got != 12 {
		t.Fatalf("Gold Mine destruction score = %d, want 12", got)
	}
}

func TestResourceCostAccessorsReturnCopies(t *testing.T) {
	building := BuildingResourceCost(BuildingBarracks)
	building[ResourceWood] = 99
	if got := BuildingResourceCost(BuildingBarracks)[ResourceWood]; got != 6 {
		t.Fatalf("mutated shared building cost: Wood = %d", got)
	}

	unit := UnitResourceCost(UnitArcher)
	unit[ResourceWood] = 99
	if got := UnitResourceCost(UnitArcher)[ResourceWood]; got != 4 {
		t.Fatalf("mutated shared unit cost: Wood = %d", got)
	}

	consumes := BuildingConsumes(BuildingBank)
	consumes[ResourceGold] = 99
	if got := BuildingConsumes(BuildingBank)[ResourceGold]; got != 5 {
		t.Fatalf("mutated shared Bank consumption: Gold = %d", got)
	}
}

func TestMineProductionByTerrain(t *testing.T) {
	tests := []struct {
		tile      TileType
		resources map[ResourceType]uint32
		coins     int32
	}{
		{TileRock, map[ResourceType]uint32{ResourceStone: 1}, 0},
		{TileIron, map[ResourceType]uint32{ResourceIron: 1}, 0},
		{TileGold, map[ResourceType]uint32{ResourceGold: 1}, 2},
		{TileCoal, nil, 0},
	}

	for _, test := range tests {
		t.Run(test.tile.String(), func(t *testing.T) {
			if got := BuildingProduces(BuildingMine, test.tile); !reflect.DeepEqual(got, test.resources) {
				t.Errorf("BuildingProduces() = %v, want %v", got, test.resources)
			}
			if got := BuildingCoinsProduces(BuildingMine, test.tile); got != test.coins {
				t.Errorf("BuildingCoinsProduces() = %d, want %d", got, test.coins)
			}
		})
	}
}

func TestFactionControlScore(t *testing.T) {
	m := Map{Grid: [][]Cell{
		{
			{Tile: TilePlains, Owner: 0, Building: &BuildingData{Type: BuildingTownhall}},
			{Tile: TilePlains, Owner: 0, Building: &BuildingData{Type: BuildingFarm}},
		},
		{
			{Tile: TileGold, Owner: 0, Building: &BuildingData{Type: BuildingMine}},
			{Tile: TilePlains, Owner: 1, Building: &BuildingData{Type: BuildingBarracks}},
		},
	}}

	if got := FactionControlScore(&m, 0); got != 7 {
		t.Fatalf("FactionControlScore() = %d, want 7", got)
	}
}
