package game

import "testing"

func TestFactionRoundIncome(t *testing.T) {
	m := Map{
		Grid: [][]Cell{
			{{Tile: TilePlains, Owner: 0, Building: &BuildingData{Type: BuildingTownhall, HP: 20}}},
			{{Tile: TilePlains, Owner: 0, Building: &BuildingData{Type: BuildingFarm, HP: 8}}},
			{{Tile: TileForest, Owner: 0, Building: &BuildingData{Type: BuildingForester, HP: 8}}},
			{{Tile: TileGold, Owner: 0, Building: &BuildingData{Type: BuildingMine, HP: 10}}},
			{{Tile: TilePlains, Owner: 0, Building: &BuildingData{Type: BuildingBank, HP: 10}}},
		},
	}

	coins, resources := FactionRoundIncome(&m, 0)

	if coins != 2 {
		t.Fatalf("coins = %d, want 2", coins)
	}
	if resources[ResourceFood] != 1 || resources[ResourceWood] != 1 {
		t.Fatalf("resources = %v, want Food 1 and Wood 1", resources)
	}
	if resources[ResourceGold] != 1 {
		t.Fatalf("Gold Mine production = %v, want Gold 1", resources)
	}
}

func TestResolveFactionRoundIncomeRunsFundedBanksWithoutUnderflow(t *testing.T) {
	m := Map{Grid: [][]Cell{
		{{Tile: TilePlains, Owner: 0, Building: &BuildingData{Type: BuildingTownhall}}},
		{{Tile: TilePlains, Owner: 0, Building: &BuildingData{Type: BuildingBank}}},
		{{Tile: TilePlains, Owner: 0, Building: &BuildingData{Type: BuildingBank}}},
	}}

	current := Resources{ResourceGold: 4, ResourceFood: 2}
	coins, resources := ResolveFactionRoundIncome(&m, 0, current)
	if coins != 1 {
		t.Fatalf("unfunded Bank coins = %d, want Townhall-only income 1", coins)
	}
	if resources[ResourceGold] != 4 {
		t.Fatalf("unfunded Bank Gold = %d, want 4", resources[ResourceGold])
	}
	if current[ResourceGold] != 4 {
		t.Fatalf("income mutated input Gold to %d", current[ResourceGold])
	}

	current[ResourceGold] = 10
	coins, resources = ResolveFactionRoundIncome(&m, 0, current)
	if coins != 11 {
		t.Fatalf("two funded Bank coins = %d, want 11", coins)
	}
	if resources[ResourceGold] != 0 {
		t.Fatalf("two funded Bank Gold = %d, want 0", resources[ResourceGold])
	}
}

func TestResolveFactionRoundIncomeProductionCanFundBankSameRound(t *testing.T) {
	m := Map{Grid: [][]Cell{
		{{Tile: TileGold, Owner: 0, Building: &BuildingData{Type: BuildingMine}}},
		{{Tile: TilePlains, Owner: 0, Building: &BuildingData{Type: BuildingBank}}},
	}}

	coins, resources := ResolveFactionRoundIncome(
		&m,
		0,
		Resources{ResourceGold: 4},
	)
	if coins != 6 {
		t.Fatalf("Gold Mine plus funded Bank coins = %d, want 6", coins)
	}
	if resources[ResourceGold] != 0 {
		t.Fatalf("same-round produced Gold was not consumed: %v", resources)
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
