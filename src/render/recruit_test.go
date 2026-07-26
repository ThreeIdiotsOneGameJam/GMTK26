package render

import (
	"testing"

	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

func TestRecruitmentOntoBuildingConsumesWorldClick(t *testing.T) {
	source := game.NewHex(0, 0)
	target := game.NewHex(1, 0)
	m := game.Map{
		Grid: [][]game.Cell{
			{
				{
					Tile:     game.TilePlains,
					Owner:    0,
					Building: game.BuildingBarracks,
				},
			},
			{
				{
					Tile:     game.TilePlains,
					Owner:    0,
					Building: game.BuildingFarm,
				},
			},
		},
		GridSize: vec.Vec2i{X: 2, Y: 1},
	}
	recruited := false
	r := WorldRenderer{
		ActionsEnabled: true,
		LocalFaction:   0,
		SelectedHex:    &source,
		SelectedKind:   SelectionBuilding,
		RecruitToPlace: game.UnitPeasant,
		OnRecruit: func(from, to game.Hex, unit game.UnitType) bool {
			recruited = from == source && to == target && unit == game.UnitPeasant
			return true
		},
	}

	if !r.updateRecruitPlacement(&m, target, true) {
		t.Fatal("successful recruitment did not consume the world click")
	}
	if !recruited {
		t.Fatal("recruitment callback was not invoked")
	}
	if r.SelectedHex != nil || r.SelectedKind != SelectionNone {
		t.Fatal("destination building was selected after recruitment")
	}
}
