package render

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/global"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/rlvec"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

type buildingPreview struct {
	Type    game.BuildingType
	Hex     game.Hex
	Tint    color.RGBA
	Visible bool
}

func (r *WorldRenderer) updateBuildingPlacement(m *game.Map, hex game.Hex) {
	r.buildingPreview.Visible = false

	if global.UIBlocksWorldInput ||
		r.BuildingToPlace == game.BuildingUnknown ||
		!m.HexInsideBounds(hex) {
		return
	}

	canPlace := game.BuildingCanPlace(m, r.BuildingToPlace, hex)
	if rl.IsMouseButtonPressed(rl.MouseButtonLeft) && canPlace {
		if r.OnPlaceBuilding == nil || !r.OnPlaceBuilding(hex, r.BuildingToPlace) {
			m.GetCell(hex).Building = r.BuildingToPlace
		}
		canPlace = false
	}

	tint := rl.Red
	if canPlace {
		tint = rl.Green
	}
	tint.A = 123

	r.buildingPreview = buildingPreview{
		Type:    r.BuildingToPlace,
		Hex:     hex,
		Tint:    tint,
		Visible: true,
	}
}

func getBuildingRect(building game.BuildingType, hovered bool) rl.Rectangle {
	y := float32(0.0)
	if hovered {
		y = 96.0
	}

	switch building {
	case game.BuildingUnknown:
		return rl.Rectangle{}
	case game.BuildingTownhall:
		return rl.Rectangle{X: 96.0 * 4.0, Y: y, Width: 96.0, Height: 96.0}
	case game.BuildingBarracks:
		return rl.Rectangle{X: 96.0 * 3.0, Y: y, Width: 96.0, Height: 96.0}
	case game.BuildingFarm:
		return rl.Rectangle{X: 96.0, Y: y, Width: 96.0, Height: 96.0}
	case game.BuildingForester:
		return rl.Rectangle{X: 96.0 * 2.0, Y: y, Width: 96.0, Height: 96.0}
	case game.BuildingMine:
		return rl.Rectangle{X: 0.0, Y: y, Width: 96.0, Height: 96.0}
	}
	return rl.Rectangle{X: 0.0, Y: y, Width: 96.0, Height: 96.0}
}

func (r *WorldRenderer) drawBuildings(m *game.Map, visible []visibleTile, mousePos vec.Vec2) {
	mouseHex := r.PixelToHex(mousePos)

	draw := func(worldPos vec.Vec2, building game.BuildingType, tint color.RGBA, hovered bool) {
		if building == game.BuildingUnknown {
			return
		}
		col := rl.ColorLerp(rl.White, tint, 0.5)

		s := vec.Vec2{X: 96, Y: 96}
		p := worldPos.Sub(s.Div(vec.Vec2{X: 2, Y: 2}))
		rl.DrawTextureRec(r.buildingsTexture, getBuildingRect(building, hovered), rlvec.ToRL(p), col)
	}

	for _, b := range visible {
		building := m.Grid[b.hex.X][b.hex.Y].Building
		draw(b.position, building, rl.White, mouseHex == b.hex)
	}

	if r.buildingPreview.Visible {
		worldPos := r.HexToPixel(r.buildingPreview.Hex.Vec2i)
		draw(worldPos, r.buildingPreview.Type, r.buildingPreview.Tint, false)
	}
}
