package world

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
	v "github.com/threeidiotsonegamejam/gmtk26/src/mathutil/vec"
)

type TileDrawState int32

const (
	DrawStateBegin TileDrawState = iota
	DrawStateNormal
	DrawStateEnd
)

type TileData struct {
	Type  string
	Color color.RGBA
}

type Tile interface {
	Data() TileData
	Tick()
	Update(w World, delta float32, pos v.Vec2)
	Draw(world World, pos v.Vec2, tile v.Vec2i, state TileDrawState)
}

type VoidTile struct{}

func (t *VoidTile) Data() TileData {
	return TileData{
		Type:  "void",
		Color: color.RGBA{R: 0, G: 0, B: 0, A: 0},
	}
}
func (t *VoidTile) Tick()                                     {}
func (t *VoidTile) Update(w World, delta float32, pos v.Vec2) {}
func (t *VoidTile) Draw(world World, pos v.Vec2, tile v.Vec2i, state TileDrawState) {
	if state == DrawStateBegin {
		rl.BeginShaderMode(world.VoidShader)
		rl.Begin(rl.Triangles)
	}
	x := pos.X
	y := pos.Y
	size := world.HexSize
	w := size.X * 2.0
	h := size.Y * sqrt3
	wp := w / 4.0
	hp := h / 2.0
	ox := w / 2.0
	oy := hp

	a := v.Vec2{X: x - ox + wp, Y: y - oy}
	b := v.Vec2{X: x - ox, Y: y - oy + hp}
	c := v.Vec2{X: x - ox + wp, Y: y - oy + h}
	d := v.Vec2{X: x - ox + wp*3, Y: y - oy + h}
	e := v.Vec2{X: x - ox + w, Y: y - oy + hp}
	f := v.Vec2{X: x - ox + wp*3, Y: y - oy}
	center := v.Vec2{X: x, Y: y}

	data := world.GetTile(tile).Data()
	neighbors := world.GetNeighbors(tile)

	isEdge := func(t Tile) bool {
		if t == nil {
			return false
		}
		return t.Data().Type != data.Type
	}
	drawSection := func(v1, v2 v.Vec2, b1, b2 bool, edge bool) {
		if (b1 || b2) && !edge {
			mid := v1.Lerp(v2, 0.5)
			if b1 {
				rl.Color4ub(255, 0, 0, 255)
			} else {
				rl.Color4ub(0, 0, 0, 0)
			}
			rl.Vertex2f(v1.X, v1.Y)

			rl.Color4ub(0, 0, 0, 0)
			rl.Vertex2f(mid.X, mid.Y)

			rl.Color4ub(0, 0, 0, 0)
			rl.Vertex2f(center.X, center.Y)

			rl.Color4ub(0, 0, 0, 0)
			rl.Vertex2f(mid.X, mid.Y)

			if b2 {
				rl.Color4ub(255, 0, 0, 255)
			} else {
				rl.Color4ub(0, 0, 0, 0)
			}
			rl.Vertex2f(v2.X, v2.Y)

			rl.Color4ub(0, 0, 0, 0)
			rl.Vertex2f(center.X, center.Y)

			return
		}

		if edge {
			rl.Color4ub(255, 0, 0, 255)
		} else {
			rl.Color4ub(0, 0, 0, 0)
		}
		rl.Vertex2f(v1.X, v1.Y)

		if edge {
			rl.Color4ub(255, 0, 0, 255)
		} else {
			rl.Color4ub(0, 0, 0, 0)
		}
		rl.Vertex2f(v2.X, v2.Y)

		rl.Color4ub(0, 0, 0, 0)
		rl.Vertex2f(center.X, center.Y)
	}

	drawSection(a, b, isEdge(neighbors.N), isEdge(neighbors.SW), isEdge(neighbors.NW))
	drawSection(b, c, isEdge(neighbors.NW), isEdge(neighbors.S), isEdge(neighbors.SW))
	drawSection(c, d, isEdge(neighbors.SW), isEdge(neighbors.SE), isEdge(neighbors.S))
	drawSection(d, e, isEdge(neighbors.S), isEdge(neighbors.NE), isEdge(neighbors.SE))
	drawSection(e, f, isEdge(neighbors.SE), isEdge(neighbors.N), isEdge(neighbors.NE))
	drawSection(f, a, isEdge(neighbors.NE), isEdge(neighbors.NW), isEdge(neighbors.N))

	if state == DrawStateEnd {
		rl.End()
		rl.EndShaderMode()
	}
}

type UnkownTile struct{}

func (t *UnkownTile) Data() TileData {
	return TileData{
		Type:  "unkown",
		Color: color.RGBA{R: 211, G: 191, B: 139, A: 255},
	}
}
func (t *UnkownTile) Tick()                                                           {}
func (t *UnkownTile) Update(w World, delta float32, pos v.Vec2)                       {}
func (t *UnkownTile) Draw(world World, pos v.Vec2, tile v.Vec2i, state TileDrawState) {}

// These tiles are temporary
type WaterTile struct{}

func (t *WaterTile) Data() TileData {
	return TileData{
		Type:  "water",
		Color: color.RGBA{R: 50, G: 70, B: 200, A: 255},
	}
}
func (t *WaterTile) Tick()                                                           {}
func (t *WaterTile) Update(w World, delta float32, pos v.Vec2)                       {}
func (t *WaterTile) Draw(world World, pos v.Vec2, tile v.Vec2i, state TileDrawState) {}

type GrassTile struct{}

func (t *GrassTile) Data() TileData {
	return TileData{
		Type:  "grass",
		Color: color.RGBA{R: 63, G: 200, B: 140, A: 255},
	}
}
func (t *GrassTile) Tick()                                                           {}
func (t *GrassTile) Update(w World, delta float32, pos v.Vec2)                       {}
func (t *GrassTile) Draw(world World, pos v.Vec2, tile v.Vec2i, state TileDrawState) {}
