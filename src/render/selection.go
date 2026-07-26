package render

import (
	"fmt"
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

type SelectionKind uint8

const (
	SelectionNone SelectionKind = iota
	SelectionTroop
	SelectionBuilding
)

const (
	selectionMenuWidth     float32 = 200
	selectionMenuPadding   float32 = 5
	selectionMenuRowHeight float32 = 32
	selectionMenuGap       float32 = 4
)

type selectionMenu struct {
	Hex      game.Hex
	Position vec.Vec2
	Visible  bool
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
	r.RecruitToPlace = game.TroopUnknown
	r.buildingPreview.Visible = false
	r.clearSelection()
}

func (r *WorldRenderer) selectCell(hex game.Hex, cell *game.Cell) {
	if cell == nil {
		r.clearSelection()
		return
	}
	hasTroop := cell.Troop != game.TroopUnknown
	hasBuilding := cell.Building != game.BuildingUnknown
	switch {
	case hasTroop && hasBuilding:
		r.clearSelection()
		r.selectionMenu = selectionMenu{
			Hex:      hex,
			Position: r.MousePosition.Add(vec.Vec2{X: 10, Y: 10}),
			Visible:  true,
		}
	case hasTroop:
		r.selectAt(hex, SelectionTroop)
	case hasBuilding:
		r.selectAt(hex, SelectionBuilding)
	default:
		r.clearSelection()
	}
}

// updateSelectionMenu returns true only when it consumes a left click. Right
// mouse input remains exclusively available for map panning.
func (r *WorldRenderer) updateSelectionMenu(m *game.Map) bool {
	if !r.selectionMenu.Visible {
		return false
	}
	if !rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
		return false
	}

	cell := m.GetCell(r.selectionMenu.Hex)
	row := r.selectionMenuRowAt(r.MousePosition)
	switch row {
	case 0:
		if cell != nil && cell.Troop != game.TroopUnknown {
			r.selectAt(r.selectionMenu.Hex, SelectionTroop)
		} else {
			r.clearSelection()
		}
	case 1:
		if cell != nil && cell.Building != game.BuildingUnknown {
			r.selectAt(r.selectionMenu.Hex, SelectionBuilding)
		} else {
			r.clearSelection()
		}
	default:
		r.selectionMenu.Visible = false
	}
	return true
}

func (r *WorldRenderer) selectionMenuBounds() rl.Rectangle {
	height := selectionMenuPadding*2 + selectionMenuRowHeight*2 + selectionMenuGap
	x := r.selectionMenu.Position.X
	y := r.selectionMenu.Position.Y
	if r.viewport.Texture.Width > 0 {
		x = min(x, float32(r.viewport.Texture.Width)-selectionMenuWidth-selectionMenuPadding)
	}
	if r.viewport.Texture.Height > 0 {
		y = min(y, float32(r.viewport.Texture.Height)-height-selectionMenuPadding)
	}
	x = max(selectionMenuPadding, x)
	y = max(selectionMenuPadding, y)
	return rl.Rectangle{X: x, Y: y, Width: selectionMenuWidth, Height: height}
}

func (r *WorldRenderer) selectionMenuRowAt(point vec.Vec2) int {
	bounds := r.selectionMenuBounds()
	x := point.X
	y := point.Y
	if x < bounds.X+selectionMenuPadding ||
		x > bounds.X+bounds.Width-selectionMenuPadding {
		return -1
	}
	firstY := bounds.Y + selectionMenuPadding
	if y >= firstY && y <= firstY+selectionMenuRowHeight {
		return 0
	}
	secondY := firstY + selectionMenuRowHeight + selectionMenuGap
	if y >= secondY && y <= secondY+selectionMenuRowHeight {
		return 1
	}
	return -1
}

func (r *WorldRenderer) drawSelectionMenu(m *game.Map) {
	if !r.selectionMenu.Visible {
		return
	}
	cell := m.GetCell(r.selectionMenu.Hex)
	if cell == nil {
		return
	}

	bounds := r.selectionMenuBounds()
	rl.DrawRectangleRec(bounds, color.RGBA{R: 24, G: 28, B: 34, A: 245})
	rl.DrawRectangleLinesEx(bounds, 1, color.RGBA{R: 220, G: 225, B: 232, A: 220})

	labels := [2]string{
		fmt.Sprintf("Troop: %s", cell.Troop),
		fmt.Sprintf("Building: %s", cell.Building),
	}
	firstY := bounds.Y + selectionMenuPadding
	for i, label := range labels {
		y := firstY + float32(i)*(selectionMenuRowHeight+selectionMenuGap)
		row := rl.Rectangle{
			X:      bounds.X + selectionMenuPadding,
			Y:      y,
			Width:  bounds.Width - selectionMenuPadding*2,
			Height: selectionMenuRowHeight,
		}
		fill := color.RGBA{R: 49, G: 57, B: 68, A: 250}
		if r.selectionMenuRowAt(r.MousePosition) == i {
			fill = color.RGBA{R: 74, G: 96, B: 118, A: 255}
		}
		rl.DrawRectangleRec(row, fill)
		rl.DrawText(label, int32(row.X+8), int32(row.Y+7), 18, rl.White)
	}
}
