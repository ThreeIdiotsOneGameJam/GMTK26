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
	if resources[ResourceFood] != 1 || resources[ResourceWood] != 1 {
		t.Fatalf("resources = %v, want Food 1 and Wood 1", resources)
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
			{{Tile: TileForest, Owner: 0, Building: &BuildingData{Type: BuildingForester, HP: 8}}},
		},
	}

	if !CanAffordUnitAfterRoundIncome(
		&m,
		0,
		UnitPeasant,
		7,
		Resources{ResourceFood: 3},
	) {
		t.Fatal("incoming Townhall coin and Farm food did not fund Peasant")
	}
	if CanAffordUnitAfterRoundIncome(
		&m,
		0,
		UnitArcher,
		13,
		Resources{ResourceFood: 5, ResourceWood: 2},
	) {
		t.Fatal("Archer was affordable without enough incoming Wood")
	}
	if !CanAffordUnitAfterRoundIncome(
		&m,
		0,
		UnitArcher,
		13,
		Resources{ResourceFood: 5, ResourceWood: 3},
	) {
		t.Fatal("current and incoming resources did not combine to fund Archer")
	}
}

func TestCanAffordBarracksAfterRoundIncome(t *testing.T) {
	m := Map{
		Grid: [][]Cell{
			{{Tile: TilePlains, Owner: 0, Building: &BuildingData{Type: BuildingTownhall}}},
			{{Tile: TileForest, Owner: 0, Building: &BuildingData{Type: BuildingForester}}},
			{{Tile: TileRock, Owner: 0, Building: &BuildingData{Type: BuildingMine}}},
		},
	}
	resources := Resources{ResourceWood: 5, ResourceStone: 3}

	if !CanAffordBuildingAfterRoundIncome(
		&m,
		0,
		BuildingBarracks,
		23,
		resources,
	) {
		t.Fatal("incoming Townhall, Forester, and Mine income did not fund Barracks")
	}
	resources[ResourceStone] = 2
	if CanAffordBuildingAfterRoundIncome(
		&m,
		0,
		BuildingBarracks,
		23,
		resources,
	) {
		t.Fatal("Barracks was affordable without enough current and incoming Stone")
	}
}
