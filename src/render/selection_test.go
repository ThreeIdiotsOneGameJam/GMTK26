package render

import (
	"testing"

	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

func TestSelectCellWithTroopAndBuildingOpensContextMenuOnly(t *testing.T) {
	r := WorldRenderer{MousePosition: vec.Vec2{X: 40, Y: 50}}
	hex := game.NewHex(2, 3)
	cell := &game.Cell{
		Troop:    game.TroopScout,
		Building: game.BuildingFarm,
	}

	r.selectCell(hex, cell)

	if r.SelectedHex != nil || r.SelectedKind != SelectionNone {
		t.Fatal("mixed tile selected an object before the player chose one")
	}
	if !r.selectionMenu.Visible || r.selectionMenu.Hex != hex {
		t.Fatal("mixed tile did not open its selection context menu")
	}
}

func TestSelectCellUsesUnambiguousObjectKind(t *testing.T) {
	hex := game.NewHex(1, 1)
	tests := []struct {
		name string
		cell game.Cell
		want SelectionKind
	}{
		{
			name: "troop",
			cell: game.Cell{Troop: game.TroopPeasant},
			want: SelectionTroop,
		},
		{
			name: "building",
			cell: game.Cell{Building: game.BuildingBarracks},
			want: SelectionBuilding,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := WorldRenderer{}
			r.selectCell(hex, &tt.cell)
			if r.SelectedHex == nil || *r.SelectedHex != hex || r.SelectedKind != tt.want {
				t.Fatalf("selection = (%v, %v), want (%v, %v)", r.SelectedHex, r.SelectedKind, hex, tt.want)
			}
			if r.selectionMenu.Visible {
				t.Fatal("unambiguous tile opened a context menu")
			}
		})
	}
}

func TestClearMouseSlotClearsSelectionAndTargetingOnly(t *testing.T) {
	selected := game.NewHex(1, 1)
	order := game.MovementOrder{
		Current:     selected,
		Destination: game.NewHex(3, 3),
	}
	r := WorldRenderer{
		ActionsEnabled:  true,
		SelectedHex:     &selected,
		SelectedKind:    SelectionTroop,
		BuildingToPlace: game.BuildingFarm,
		RecruitToPlace:  game.TroopScout,
		Orders:          []game.MovementOrder{order},
		PreviewPath:     []game.Hex{selected, game.NewHex(1, 2)},
		PreviewStops:    []game.Hex{game.NewHex(1, 2)},
		buildingPreview: buildingPreview{Visible: true},
		queuedBuilding:  buildingPreview{Visible: true},
		selectionMenu:   selectionMenu{Visible: true},
	}

	r.clearMouseSlot()

	if r.SelectedHex != nil || r.SelectedKind != SelectionNone {
		t.Fatal("active object selection was not cleared")
	}
	if r.BuildingToPlace != game.BuildingUnknown ||
		r.RecruitToPlace != game.TroopUnknown {
		t.Fatal("placement targeting was not cleared")
	}
	if r.buildingPreview.Visible ||
		r.selectionMenu.Visible ||
		r.PreviewPath != nil ||
		r.PreviewStops != nil {
		t.Fatal("mouse-slot state was not cleared")
	}
	if !r.queuedBuilding.Visible {
		t.Fatal("middle click cancelled the pending building action")
	}
	if len(r.Orders) != 1 || r.Orders[0] != order {
		t.Fatal("middle click changed a queued movement order")
	}
}
