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
		r.drawUnit(v.position, unit.Type, factionColors[unit.Owner])
		if unit.Owner == r.LocalFaction {
			drawHPBar(v.position, r.HexSize, unit.HP, game.GetUnitStats(unit.Type).MaxHP)
		}
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
