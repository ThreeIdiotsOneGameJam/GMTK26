package render

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

func (r *WorldRenderer) drawUnits(m *game.Map, visible []visibleTile) {
	for _, v := range visible {
		cell := &m.Grid[v.hex.X][v.hex.Y]
		if cell.Unit == game.UnitUnknown {
			continue
		}

		owner := cell.UnitOwner
		if owner < 0 || int(owner) >= len(factionColors) {
			continue
		}
		if r.unitEndpointAnimating(v.hex, owner, cell.Unit) {
			continue
		}
		drawUnitMarker(v.position, r.HexSize, cell.Unit, factionColors[owner])
	}
}

func drawUnitMarker(center, hexSize vec.Vec2, unit game.UnitType, markerColor color.RGBA) {
	size := hexSize.Mul(vec.Vec2{X: 0.3, Y: 0.3})
	if unit == game.UnitScout {
		radius := min(size.X, size.Y) * 0.75
		rl.DrawPoly(rl.Vector2(center), 4, radius, 45, markerColor)
		rl.DrawPolyLines(rl.Vector2(center), 4, radius, 45, rl.White)
		return
	}
	v1, v2, v3 := unitTriangle(center, size, unit)
	rl.DrawTriangle(v1, v2, v3, markerColor)
}

func unitTriangle(center vec.Vec2, size vec.Vec2, unit game.UnitType) (rl.Vector2, rl.Vector2, rl.Vector2) {
	hw := size.X
	hh := size.Y

	var tip, bl, br vec.Vec2

	switch unit {
	case game.UnitArcher:
		tip = vec.Vec2{X: center.X + hw, Y: center.Y}
		bl = vec.Vec2{X: center.X - hw, Y: center.Y - hh}
		br = vec.Vec2{X: center.X - hw, Y: center.Y + hh}
	case game.UnitKnight:
		tip = vec.Vec2{X: center.X, Y: center.Y + hh}
		bl = vec.Vec2{X: center.X + hw, Y: center.Y - hh}
		br = vec.Vec2{X: center.X - hw, Y: center.Y - hh}
	default:
		tip = vec.Vec2{X: center.X, Y: center.Y - hh}
		bl = vec.Vec2{X: center.X - hw, Y: center.Y + hh}
		br = vec.Vec2{X: center.X + hw, Y: center.Y + hh}
	}

	return rl.Vector2(tip), rl.Vector2(bl), rl.Vector2(br)
}
