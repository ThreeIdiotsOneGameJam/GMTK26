package world

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/util"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

type TileDrawState int32

const (
	DrawStateBegin TileDrawState = iota
	DrawStateNormal
	DrawStateEnd
)

type Tile struct {
	Type  string
	Color color.RGBA
	Draw  func(w World, hex vec.Vec2i, worldPos vec.Vec2, state TileDrawState)
}

var VoidTile = Tile{
	Type:  "void",
	Color: color.RGBA{R: 0, G: 0, B: 0, A: 0},
	Draw: func(w World, hex vec.Vec2i, worldPos vec.Vec2, state TileDrawState) {
		t := w.GetCell(hex).Tile
		DrawEdge(t, w, hex, worldPos, state, t.Color)
	},
}

var WaterTile = Tile{
	Type:  "water",
	Color: color.RGBA{R: 50, G: 70, B: 200, A: 255},
	Draw: func(w World, hex vec.Vec2i, worldPos vec.Vec2, state TileDrawState) {
		t := w.GetCell(hex).Tile
		col := t.Color
		col = *util.ColorSub(col, 50)
		DrawEdge(t, w, hex, worldPos, state, col)
	},
}

var PlainsTile = Tile{
	Type:  "plains",
	Color: color.RGBA{R: 50, G: 200, B: 80, A: 255},
}

var ForestTile = Tile{
	Type:  "forest",
	Color: color.RGBA{R: 50, G: 150, B: 100, A: 255},
}

var DesertTile = Tile{
	Type:  "desert",
	Color: color.RGBA{R: 196, G: 190, B: 100, A: 255},
}

var JungleTile = Tile{
	Type:  "jungle",
	Color: color.RGBA{R: 50, G: 120, B: 20, A: 255},
}

var RockTile = Tile{
	Type:  "rock",
	Color: color.RGBA{R: 150, G: 150, B: 150, A: 255},
}

func DrawEdge(t Tile, w World, tile vec.Vec2i, pos vec.Vec2, state TileDrawState, col color.RGBA) {
	if state == DrawStateBegin {
		rl.BeginShaderMode(w.VoidShader)
		rl.Begin(rl.Triangles)
	}
	if state == DrawStateEnd {
		defer rl.End()
		defer rl.EndShaderMode()
	}

	neighbors := w.GetNeighbors(tile)

	isEdge := func(c *Cell) bool {
		if c == nil {
			return false
		}
		return c.Tile.Type != t.Type
	}

	if !isEdge(neighbors.N) && !isEdge(neighbors.NE) && !isEdge(neighbors.NW) && !isEdge(neighbors.S) && !isEdge(neighbors.SE) && !isEdge(neighbors.SW) {
		return
	}

	x := pos.X
	y := pos.Y
	size := w.HexSize
	width := size.X * 2.0
	height := size.Y * sqrt3
	wp := width / 4.0
	hp := height / 2.0
	ox := width / 2.0
	oy := hp

	a := vec.Vec2{X: x - ox + wp, Y: y - oy}
	b := vec.Vec2{X: x - ox, Y: y - oy + hp}
	c := vec.Vec2{X: x - ox + wp, Y: y - oy + height}
	d := vec.Vec2{X: x - ox + wp*3, Y: y - oy + height}
	e := vec.Vec2{X: x - ox + width, Y: y - oy + hp}
	f := vec.Vec2{X: x - ox + wp*3, Y: y - oy}
	center := vec.Vec2{X: x, Y: y}

	drawSection := func(v1, v2 vec.Vec2, b1, b2 bool, edge bool) {
		if (b1 || b2) && !edge {
			mid := v1.Lerp(v2, 0.5)
			if b1 {
				rl.Normal3f(1.0, 0.0, 0.0)
			} else {
				rl.Normal3f(0.0, 0.0, 0.0)
			}
			rl.Vertex2f(v1.X, v1.Y)

			rl.Normal3f(0.0, 0.0, 0.0)
			rl.Vertex2f(mid.X, mid.Y)

			rl.Normal3f(0.0, 0.0, 0.0)
			rl.Vertex2f(center.X, center.Y)

			rl.Normal3f(0.0, 0.0, 0.0)
			rl.Vertex2f(mid.X, mid.Y)

			if b2 {
				rl.Normal3f(1.0, 0.0, 0.0)
			} else {
				rl.Normal3f(0.0, 0.0, 0.0)
			}
			rl.Vertex2f(v2.X, v2.Y)

			rl.Normal3f(0.0, 0.0, 0.0)
			rl.Vertex2f(center.X, center.Y)

			return
		}

		if edge {
			rl.Normal3f(1.0, 0.0, 0.0)
		} else {
			rl.Normal3f(0.0, 0.0, 0.0)
		}
		rl.Vertex2f(v1.X, v1.Y)

		if edge {
			rl.Normal3f(1.0, 0.0, 0.0)
		} else {
			rl.Normal3f(0.0, 0.0, 0.0)
		}
		rl.Vertex2f(v2.X, v2.Y)

		rl.Normal3f(0.0, 0.0, 0.0)
		rl.Vertex2f(center.X, center.Y)
	}

	rl.Color4ub(col.R, col.G, col.B, col.A)
	rl.TexCoord2f(float32(pos.X), float32(pos.Y))
	drawSection(a, b, isEdge(neighbors.N), isEdge(neighbors.SW), isEdge(neighbors.NW))
	drawSection(b, c, isEdge(neighbors.NW), isEdge(neighbors.S), isEdge(neighbors.SW))
	drawSection(c, d, isEdge(neighbors.SW), isEdge(neighbors.SE), isEdge(neighbors.S))
	drawSection(d, e, isEdge(neighbors.S), isEdge(neighbors.NE), isEdge(neighbors.SE))
	drawSection(e, f, isEdge(neighbors.SE), isEdge(neighbors.N), isEdge(neighbors.NE))
	drawSection(f, a, isEdge(neighbors.NE), isEdge(neighbors.NW), isEdge(neighbors.N))
}
