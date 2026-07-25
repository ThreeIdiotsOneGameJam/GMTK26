package render

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/global"
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

func buildingColor(building game.BuildingType) color.RGBA {
	switch building {
	case game.BuildingUnknown:
		return color.RGBA{}
	case game.BuildingBarracks:
		return color.RGBA{R: 80, G: 150, B: 20, A: 255}
	case game.BuildingFarm:
		return color.RGBA{R: 200, G: 170, B: 20, A: 255}
	case game.BuildingForester:
		return color.RGBA{R: 140, G: 100, B: 20, A: 255}
	case game.BuildingMine:
		return color.RGBA{R: 20, G: 20, B: 20, A: 255}
	}
	return rl.Magenta
}

func (r *WorldRenderer) drawBuildings(m *game.Map, visible []visibleTile) {
	draw := func(worldPos vec.Vec2, building game.BuildingType, tint color.RGBA) {
		if building == game.BuildingUnknown {
			return
		}

		s := m.GridSize.Div(vec.Vec2i{X: 2, Y: 5}).Mul(vec.Vec2i{X: 1, Y: 3})
		p := worldPos.Vec2i().Sub(s.Div(vec.Vec2i{X: 2, Y: 3}).Mul(vec.Vec2i{X: 1, Y: 2}))
		rl.DrawRectangle(p.X, p.Y, s.X, s.Y, rl.ColorLerp(buildingColor(building), tint, 0.5))
	}

	for _, b := range visible {
		building := m.Grid[b.hex.X][b.hex.Y].Building
		draw(b.position, building, rl.White)
	}

	if r.buildingPreview.Visible {
		worldPos := r.HexToPixel(r.buildingPreview.Hex.Vec2i)
		draw(worldPos, r.buildingPreview.Type, r.buildingPreview.Tint)
	}
}
