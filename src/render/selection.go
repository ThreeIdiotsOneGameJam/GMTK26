package render

import (
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
)

type SelectionKind uint8

const (
	SelectionNone SelectionKind = iota
	SelectionUnit
	SelectionBuilding
)

type selectionMenu struct {
	Hex     game.Hex
	Visible bool
}

func (r *WorldRenderer) selectAt(hex game.Hex, kind SelectionKind) {
	h := hex
	r.SelectedHex = &h
	r.SelectedKind = kind
	r.selectionMenu.Visible = false
	r.PreviewPath = nil
	r.PreviewStops = nil
}

func (r *WorldRenderer) clearSelection() {
	r.SelectedHex = nil
	r.SelectedKind = SelectionNone
	r.selectionMenu.Visible = false
	r.PreviewPath = nil
	r.PreviewStops = nil
}

func (r *WorldRenderer) ClearSelection() {
	r.clearSelection()
}

func (r *WorldRenderer) clearMouseSlot() {
	r.BuildingToPlace = game.BuildingUnknown
	r.RecruitToPlace = game.UnitUnknown
	r.buildingPreview.Visible = false
	r.clearSelection()
}

func (r *WorldRenderer) selectCell(hex game.Hex, cell *game.Cell) {
	if cell == nil {
		r.clearSelection()
		return
	}
	hasUnit := cell.Unit != game.UnitUnknown
	hasBuilding := cell.Building != game.BuildingUnknown
	switch {
	case hasUnit && hasBuilding:
		r.clearSelection()
		r.selectionMenu = selectionMenu{
			Hex:     hex,
			Visible: true,
		}
	case hasUnit:
		r.selectAt(hex, SelectionUnit)
	case hasBuilding:
		r.selectAt(hex, SelectionBuilding)
	default:
		r.clearSelection()
	}
}

// SelectionMenuHex returns the tile currently waiting for the player to choose
// between its unit and building.
func (r *WorldRenderer) SelectionMenuHex() (game.Hex, bool) {
	return r.selectionMenu.Hex, r.selectionMenu.Visible
}

// ChooseSelectionMenuOption applies a choice made by the UI layer. The map is
// checked again because an authoritative state update may have changed the
// tile while the menu was open.
func (r *WorldRenderer) ChooseSelectionMenuOption(m *game.Map, kind SelectionKind) {
	if !r.selectionMenu.Visible {
		return
	}

	cell := m.GetCell(r.selectionMenu.Hex)
	switch kind {
	case SelectionUnit:
		if cell != nil && cell.Unit != game.UnitUnknown {
			r.selectAt(r.selectionMenu.Hex, kind)
			return
		}
	case SelectionBuilding:
		if cell != nil && cell.Building != game.BuildingUnknown {
			r.selectAt(r.selectionMenu.Hex, kind)
			return
		}
	}
	r.clearSelection()
}

func (r *WorldRenderer) DismissSelectionMenu() {
	r.selectionMenu.Visible = false
}
