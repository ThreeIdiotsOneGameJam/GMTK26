package render

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/rlvec"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

const unitTextureCellSize float32 = 96
const unitFactionMarkerRadius float32 = 16

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
		r.drawUnit(v.position, cell.Unit, factionColors[owner])
	}
}

func getUnitRect(unit game.UnitType) rl.Rectangle {
	switch unit {
	case game.UnitScout:
		return rl.Rectangle{X: 0, Y: 0, Width: unitTextureCellSize, Height: unitTextureCellSize}
	case game.UnitPeasant:
		return rl.Rectangle{X: unitTextureCellSize, Y: 0, Width: unitTextureCellSize, Height: unitTextureCellSize}
	case game.UnitArcher:
		return rl.Rectangle{X: unitTextureCellSize * 2, Y: 0, Width: unitTextureCellSize, Height: unitTextureCellSize}
	case game.UnitKnight:
		return rl.Rectangle{X: unitTextureCellSize * 3, Y: 0, Width: unitTextureCellSize, Height: unitTextureCellSize}
	default:
		return rl.Rectangle{}
	}
}

func (r *WorldRenderer) drawUnit(center vec.Vec2, unit game.UnitType, factionColor color.RGBA) {
	source := getUnitRect(unit)
	if source.Width == 0 {
		return
	}

	rl.DrawCircleV(rlvec.ToRL(center), unitFactionMarkerRadius, factionColor)

	size := vec.Vec2{X: unitTextureCellSize, Y: unitTextureCellSize}
	position := center.Sub(size.Mul(vec.Vec2{X: 0.5, Y: 0.5}))
	rl.DrawTextureRec(r.unitsTexture, source, rlvec.ToRL(position), rl.White)
}
