package game

import "testing"

func TestFactionRoundIncome(t *testing.T) {
	m := Map{
		Grid: [][]Cell{
			{{Tile: TilePlains, Owner: 0, Building: &BuildingData{Type: BuildingTownhall, HP: 20}}},
			{{Tile: TilePlains, Owner: 0, Building: &BuildingData{Type: BuildingFarm, HP: 8}}},
			{{Tile: TileForest, Owner: 0, Building: &BuildingData{Type: BuildingForester, HP: 8}}},
			{{Tile: TileGold, Owner: 1, Building: &BuildingData{Type: BuildingMine, HP: 10}}},
		},
	}

	coins, resources := FactionRoundIncome(&m, 0)

	if coins != 1 {
		t.Fatalf("coins = %d, want 1", coins)
	}
	if resources[ResourceFood] != 2 || resources[ResourceWood] != 2 {
		t.Fatalf("resources = %v, want Food 2 and Wood 2", resources)
	}
	if resources[ResourceGold] != 0 {
		t.Fatalf("enemy mine contributed Gold: %v", resources)
	}
}

func TestCanAffordUnitAfterRoundIncome(t *testing.T) {
	m := Map{
		Grid: [][]Cell{
			{{Tile: TilePlains, Owner: 0, Building: &BuildingData{Type: BuildingTownhall, HP: 20}}},
			{{Tile: TilePlains, Owner: 0, Building: &BuildingData{Type: BuildingFarm, HP: 8}}},
		},
	}

	if !CanAffordUnitAfterRoundIncome(&m, 0, UnitPeasant, 9, nil) {
		t.Fatal("incoming Townhall coin and Farm food did not fund Peasant")
	}
	if CanAffordUnitAfterRoundIncome(&m, 0, UnitArcher, 19, nil) {
		t.Fatal("Archer was affordable without its third Food")
	}
	if !CanAffordUnitAfterRoundIncome(
		&m,
		0,
		UnitArcher,
		19,
		Resources{ResourceFood: 1},
	) {
		t.Fatal("current and incoming resources did not combine to fund Archer")
	}
}
