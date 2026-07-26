package game

import (
	"testing"

	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

func validationTestMap(width, height int32) Map {
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

func TestProjectedRoundFundsCopiesAndAddsIncome(t *testing.T) {
	m := validationTestMap(2, 1)
	m.Grid[0][0] = Cell{
		Tile:     TilePlains,
		Owner:    0,
		Building: &BuildingData{Type: BuildingTownhall},
	}
	m.Grid[1][0] = Cell{
		Tile:     TileForest,
		Owner:    0,
		Building: &BuildingData{Type: BuildingForester},
	}
	faction := Faction{
		Coins:     7,
		Resources: Resources{ResourceWood: 2},
	}

	projected := ProjectedRoundFunds(&m, 0, faction)
	if projected.Coins != 8 || projected.Resources[ResourceWood] != 3 {
		t.Fatalf("projected funds = %+v", projected)
	}
	projected.Resources[ResourceWood] = 99
	if faction.Resources[ResourceWood] != 2 {
		t.Fatal("projected resource map aliases faction resources")
	}
}

func TestValidateBuildActionUsesExactCosts(t *testing.T) {
	m := validationTestMap(2, 1)
	m.Grid[0][0].Units = []UnitData{{Type: UnitScout, Owner: 0, HP: 3}}
	payload := BuildActionPayload{
		From:     NewHex(0, 0),
		To:       NewHex(1, 0),
		Building: BuildingBarracks,
	}

	validation := ValidateBuildAction(&m, 0, payload, Funds{
		Coins: 24,
		Resources: Resources{
			ResourceWood:  6,
			ResourceStone: 4,
		},
	})
	if !validation.Valid ||
		validation.CoinCost != 24 ||
		validation.ResourceCost[ResourceWood] != 6 ||
		validation.ResourceCost[ResourceStone] != 4 {
		t.Fatalf("validation = %+v", validation)
	}

	validation.ResourceCost[ResourceWood] = 100
	if BuildingResourceCost(BuildingBarracks)[ResourceWood] != 6 {
		t.Fatal("validation cost aliases balance data")
	}
}

func TestValidateRecruitAndAttackActions(t *testing.T) {
	m := validationTestMap(3, 1)
	m.Grid[0][0] = Cell{
		Tile:     TilePlains,
		Owner:    0,
		Building: &BuildingData{Type: BuildingBarracks},
	}
	recruit := ValidateRecruitAction(
		&m,
		0,
		RecruitActionPayload{
			From: NewHex(0, 0),
			To:   NewHex(1, 0),
			Unit: UnitArcher,
		},
		Funds{
			Coins: 14,
			Resources: Resources{
				ResourceFood: 6,
				ResourceWood: 4,
			},
		},
	)
	if !recruit.Valid {
		t.Fatalf("recruit validation = %+v", recruit)
	}

	m.Grid[0][0].Building = nil
	m.Grid[0][0].Units = []UnitData{{Type: UnitKnight, Owner: 0, HP: 8}}
	m.Grid[1][0].Units = []UnitData{{Type: UnitPeasant, Owner: 1, HP: 5}}
	attack := ValidateAdjacentAttackAction(
		&m,
		0,
		AttackActionPayload{From: NewHex(0, 0), To: NewHex(1, 0)},
	)
	if !attack.Valid {
		t.Fatalf("attack validation = %+v", attack)
	}

	m.Grid[0][0].Units[0].Type = UnitScout
	attack = ValidateAdjacentAttackAction(
		&m,
		0,
		AttackActionPayload{From: NewHex(0, 0), To: NewHex(1, 0)},
	)
	if attack.Valid {
		t.Fatal("Scout attack unexpectedly validated")
	}
}
