package render

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

func (r *WorldRenderer) drawTroops(m *game.Map, visible []visibleTile) {
	for _, v := range visible {
		cell := &m.Grid[v.hex.X][v.hex.Y]
		if cell.Troop == game.TroopUnknown {
			continue
		}

		owner := cell.TroopOwner
		if owner < 0 || int(owner) >= len(factionColors) {
			continue
		}
		if r.troopEndpointAnimating(v.hex, owner, cell.Troop) {
			continue
		}
		drawTroopMarker(v.position, r.HexSize, cell.Troop, factionColors[owner])
	}
}

func drawTroopMarker(center, hexSize vec.Vec2, troop game.TroopType, markerColor color.RGBA) {
	size := hexSize.Mul(vec.Vec2{X: 0.3, Y: 0.3})
	if troop == game.TroopScout {
		radius := min(size.X, size.Y) * 0.75
		rl.DrawPoly(rl.Vector2(center), 4, radius, 45, markerColor)
		rl.DrawPolyLines(rl.Vector2(center), 4, radius, 45, rl.White)
		return
	}
	v1, v2, v3 := troopTriangle(center, size, troop)
	rl.DrawTriangle(v1, v2, v3, markerColor)
}

func troopTriangle(center vec.Vec2, size vec.Vec2, troop game.TroopType) (rl.Vector2, rl.Vector2, rl.Vector2) {
	hw := size.X
	hh := size.Y

	var tip, bl, br vec.Vec2

	switch troop {
	case game.TroopArcher:
		tip = vec.Vec2{X: center.X + hw, Y: center.Y}
		bl = vec.Vec2{X: center.X - hw, Y: center.Y - hh}
		br = vec.Vec2{X: center.X - hw, Y: center.Y + hh}
	case game.TroopKnight:
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
