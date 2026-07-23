package world

import (
	"math"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
	v "github.com/threeidiotsonegamejam/gmtk26/src/math"
)

type World struct {
	Grid      [][]Tile
	Camera    rl.Camera2D
	HexSize   v.Vec2
	HasInit   bool
	BGShader  rl.Shader
	BGTimeLoc int32
}

var sqrt3 = float32(math.Sqrt(3.0))

func (w *World) Init() {
	w.HasInit = true

	if w.Camera.Zoom == 0.0 {
		w.Camera.Zoom = 1.0
	}
	if w.HexSize == (v.Vec2{}) {
		w.HexSize = v.Vec2{X: 48.0, Y: 36.0}
	}

	w.Grid = make([][]Tile, 32)

	for x := range len(w.Grid) {
		w.Grid[x] = make([]Tile, 24)
		for y := range len(w.Grid[x]) {
			v := rand.Float32()
			if v > 0.6 {
				w.Grid[x][y] = &VoidTile{}
			} else if v > 0.4 {
				w.Grid[x][y] = &WaterTile{}
			} else if v > 0.2 {
				w.Grid[x][y] = &GrassTile{}
			} else {
				w.Grid[x][y] = &UnkownTile{}
			}
		}
	}

	// FIXME: DEATH THIS IS DEATH!!! WELL, AT LEAST UNTIL WE ADD SOMETHING TO UNLOAD IT...
	w.BGShader = rl.LoadShader("assets/shaders/bg.vert", "assets/shaders/bg.frag")
	w.BGTimeLoc = rl.GetLocationUniform(w.BGShader.ID, "time")
}

func (w World) Update(delta float32) {

}

func (w World) Draw() {
	if rl.IsShaderValid(w.BGShader) {
		rl.SetShaderValue(w.BGShader, w.BGTimeLoc, []float32{float32(rl.GetTime())}, rl.ShaderUniformFloat)
		rl.BeginShaderMode(w.BGShader)
		rl.Begin(rl.Triangles)

		width, height := float32(rl.GetRenderWidth()), float32(rl.GetRenderHeight())

		rl.Color4ub(255, 255, 0, 255)
		rl.Normal3f(0.0, 0.0, 1.0)

		rl.TexCoord2f(0.0, 0.0)
		rl.Vertex2f(0, 0)

		rl.TexCoord2f(width, height)
		rl.Vertex2f(width, height)

		rl.TexCoord2f(width, 0.0)
		rl.Vertex2f(width, 0)

		rl.TexCoord2f(0.0, height)
		rl.Vertex2f(0, height)

		rl.TexCoord2f(width, height)
		rl.Vertex2f(width, height)

		rl.TexCoord2f(0.0, 0.0)
		rl.Vertex2f(0, 0)

		rl.End()
		rl.EndShaderMode()
	}

	rl.BeginMode2D(w.Camera)

	rlMouse := rl.GetScreenToWorld2D(rl.GetMousePosition(), w.Camera)
	mp := v.Vec2{X: rlMouse.X, Y: rlMouse.Y}

	width := w.HexSize.X * 2.0
	height := w.HexSize.Y * sqrt3

	for x := range len(w.Grid) {
		for y, tile := range w.Grid[x] {
			switch tile.(type) {
			case *VoidTile:
				continue
			}

			hex := w.PixelToHex(mp)

			tileColor := tile.Color()
			if hex.X == int32(x) && hex.Y == int32(y) {
				tileColor = rl.ColorLerp(tileColor, rl.White, 0.2)
			}

			yOffset := float32(height/2.0) * float32(x%2)
			DrawHexagon(float32(x)*width/4.0*3.0, float32(y)*height+yOffset, w.HexSize, tileColor)

			tile.Draw()
		}
	}

	rl.EndMode2D()
}

func DrawHexagon(x float32, y float32, size v.Vec2, color rl.Color) {
	w := size.X * 2.0
	h := size.Y * sqrt3
	wp := w / 4.0
	hp := h / 2.0
	ox := w / 2.0
	oy := hp

	a := rl.Vector2{X: x - ox + wp, Y: y - oy}
	b := rl.Vector2{X: x - ox, Y: y - oy + hp}
	c := rl.Vector2{X: x - ox + wp, Y: y - oy + h}
	d := rl.Vector2{X: x - ox + wp*3, Y: y - oy + h}
	e := rl.Vector2{X: x - ox + w, Y: y - oy + hp}
	f := rl.Vector2{X: x - ox + wp*3, Y: y - oy}
	center := rl.Vector2{X: x, Y: y}

	rl.Begin(rl.Triangles)
	rl.Color4ub(color.R, color.G, color.B, color.A)
	rl.Normal3f(0.0, 0.0, 1.0)

	rl.Vertex2f(a.X, a.Y)
	rl.Vertex2f(b.X, b.Y)
	rl.Vertex2f(center.X, center.Y)

	rl.Vertex2f(b.X, b.Y)
	rl.Vertex2f(c.X, c.Y)
	rl.Vertex2f(center.X, center.Y)

	rl.Vertex2f(c.X, c.Y)
	rl.Vertex2f(d.X, d.Y)
	rl.Vertex2f(center.X, center.Y)

	rl.Vertex2f(d.X, d.Y)
	rl.Vertex2f(e.X, e.Y)
	rl.Vertex2f(center.X, center.Y)

	rl.Vertex2f(e.X, e.Y)
	rl.Vertex2f(f.X, f.Y)
	rl.Vertex2f(center.X, center.Y)

	rl.Vertex2f(f.X, f.Y)
	rl.Vertex2f(a.X, a.Y)
	rl.Vertex2f(center.X, center.Y)

	rl.End()
}
