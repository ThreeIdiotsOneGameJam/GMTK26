package render

import (
	"testing"

	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

func TestClearPlacementSelectionClosesPlacementAndSourceSelection(t *testing.T) {
	selected := game.NewHex(2, 3)
	r := WorldRenderer{
		SelectedHex:     &selected,
		SelectedKind:    SelectionUnit,
		BuildingToPlace: game.BuildingFarm,
		RecruitToPlace:  game.UnitScout,
		PreviewPath:     []game.Hex{selected, game.NewHex(2, 4)},
		PreviewStops:    []game.Hex{game.NewHex(2, 4)},
		buildingPreview: buildingPreview{Visible: true},
		queuedBuilding:  buildingPreview{Visible: true},
	}

	r.clearPlacementSelection()

	if r.SelectedHex != nil {
		t.Fatal("source selection remained active after placement")
	}
	if r.SelectedKind != SelectionNone {
		t.Fatal("selection kind remained active after placement")
	}
	if r.BuildingToPlace != game.BuildingUnknown || r.RecruitToPlace != game.UnitUnknown {
		t.Fatal("placement cursor remained active after placement")
	}
	if r.PreviewPath != nil || r.PreviewStops != nil || r.buildingPreview.Visible {
		t.Fatal("placement preview remained active after placement")
	}
	if !r.queuedBuilding.Visible {
		t.Fatal("pending authoritative building ghost was cleared")
	}
}

func TestCancelQueuedBuildingOnlyAtPendingTarget(t *testing.T) {
	target := game.NewHex(2, 3)
	cancelled := false
	r := WorldRenderer{
		queuedBuilding: buildingPreview{
			Hex:     target,
			Visible: true,
		},
		OnCancelBuilding: func(got game.Hex) bool {
			cancelled = got == target
			return cancelled
		},
	}

	if r.cancelQueuedBuildingAt(game.NewHex(2, 2), true) {
		t.Fatal("right click on another tile cancelled the pending build")
	}
	if !r.queuedBuilding.Visible || cancelled {
		t.Fatal("non-target click changed pending build state")
	}
	if r.cancelQueuedBuildingAt(target, false) {
		t.Fatal("right drag cancelled the pending build")
	}
	if !r.cancelQueuedBuildingAt(target, true) {
		t.Fatal("right click on pending building was not handled")
	}
	if r.queuedBuilding.Visible || !cancelled {
		t.Fatal("pending building was not cancelled")
	}
}

func TestSameTileBuildingPlacementConsumesWorldClick(t *testing.T) {
	source := game.NewHex(0, 0)
	m := game.Map{
		Grid: [][]game.Cell{{
			{
				Tile:  game.TilePlains,
				Owner: -1,
				Units: []game.UnitData{{Type: game.UnitScout, Owner: 0, HP: 3}},
			},
		}},
		GridSize: vec.Vec2i{X: 1, Y: 1},
	}
	placed := false
	r := WorldRenderer{
		ActionsEnabled:  true,
		LocalFaction:    0,
		SelectedHex:     &source,
		SelectedKind:    SelectionUnit,
		BuildingToPlace: game.BuildingBarracks,
		OnPlaceBuilding: func(from, to game.Hex, building game.BuildingType) bool {
			placed = from == source && to == source && building == game.BuildingBarracks
			return true
		},
	}

	if !r.updateBuildingPlacement(&m, source, true) {
		t.Fatal("successful same-tile placement did not consume the world click")
	}
	if !placed {
		t.Fatal("same-tile placement callback was not invoked")
	}
	if r.SelectedHex != nil || r.SelectedKind != SelectionNone {
		t.Fatal("unit was reselected after same-tile placement")
	}
}
