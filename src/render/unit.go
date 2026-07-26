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
		if !cell.HasUnits() {
			continue
		}

		unit := &cell.Units[0]
		if int(unit.Owner) >= len(factionColors) {
			continue
		}
		if r.unitEndpointAnimating(v.hex, unit.Owner, unit.Type) {
			continue
		}
		drawUnitMarker(v.position, r.HexSize, unit.Type, factionColors[unit.Owner])
		if unit.Owner == r.LocalFaction {
			drawHPBar(v.position, r.HexSize, unit.HP, game.GetUnitStats(unit.Type).MaxHP)
		}
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

func drawHPBar(center, hexSize vec.Vec2, currentHP, maxHP int8) {
	if maxHP <= 0 {
		return
	}
	ratio := float32(currentHP) / float32(maxHP)
	barW := hexSize.X * 0.5
	barH := float32(3)
	x := center.X - barW/2
	y := center.Y + hexSize.Y*0.5 + 2

	var col color.RGBA
	switch {
	case ratio > 0.66:
		col = color.RGBA{R: 80, G: 200, B: 80, A: 220}
	case ratio > 0.33:
		col = color.RGBA{R: 220, G: 200, B: 60, A: 220}
	default:
		col = color.RGBA{R: 220, G: 60, B: 60, A: 220}
	}

	rl.DrawRectangleV(rl.Vector2{X: x - 1, Y: y - 1}, rl.Vector2{X: barW + 2, Y: barH + 2}, color.RGBA{R: 28, G: 31, B: 36, A: 180})
	rl.DrawRectangleV(rl.Vector2{X: x, Y: y}, rl.Vector2{X: barW * max(ratio, 0), Y: barH}, col)
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
