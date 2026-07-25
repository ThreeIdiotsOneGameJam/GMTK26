package render

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

type ExtraBuilding struct {
	Type game.BuildingType
	Hex  game.Hex
	Tint color.RGBA
}

func (r *WorldRenderer) DrawBuilding(m *game.Map, building game.BuildingType, hex game.Hex, tint color.RGBA) {
	r.ExtraBuildings = append(r.ExtraBuildings, ExtraBuilding{
		Type: building,
		Hex:  hex,
		Tint: tint,
	})
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

func (r *WorldRenderer) drawBuildings(m *game.Map, visible []visibleTile, extra []ExtraBuilding) {
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

	for _, b := range extra {
		worldPos := r.HexToPixel(b.Hex.Vec2i)
		draw(worldPos, b.Type, b.Tint)
	}
}
