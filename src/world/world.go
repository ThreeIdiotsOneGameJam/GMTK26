package world

import (
	"math"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
	v "github.com/threeidiotsonegamejam/gmtk26/src/math"
)

type World struct {
	Grid    [][]Tile
	Camera  rl.Camera2D
	HexSize float32
}

var sqrt3 = float32(math.Sqrt(3.0))

func (w *World) Init() {
	if w.Camera.Zoom == 0.0 {
		w.Camera.Zoom = 1.0
	}
	if w.HexSize == 0.0 {
		w.HexSize = 48.0
	}

	w.Grid = make([][]Tile, 32)

	for x := range len(w.Grid) {
		w.Grid[x] = make([]Tile, 12)
		for y := range len(w.Grid[x]) {
			v := rand.Float32()
			if v > 0.8 {
				w.Grid[x][y] = &VoidTile{}
			} else if v > 0.6 {
				w.Grid[x][y] = &WaterTile{}
			} else if v > 0.4 {
				w.Grid[x][y] = &GrassTile{}
			} else {
				w.Grid[x][y] = &UnkownTile{}
			}
		}
	}
}

func (w World) Update(delta float32) {

}

func (w World) Draw() {
	rl.BeginMode2D(w.Camera)

	rlMouse := rl.GetScreenToWorld2D(rl.GetMousePosition(), w.Camera)
	mp := v.Vec2{X: rlMouse.X, Y: rlMouse.Y}

	width := w.HexSize * 2.0
	height := w.HexSize * sqrt3

	for x := range len(w.Grid) {
		for y, tile := range w.Grid[x] {

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

func DrawHexagon(x float32, y float32, r float32, color rl.Color) {
	w := r * 2.0
	h := r * sqrt3
	wp := w / 4.0
	hp := h / 2.0
	ox := w / 2.0
	oy := hp

	rl.DrawTriangle(
		rl.Vector2{X: x - ox + wp, Y: y - oy},
		rl.Vector2{X: x - ox, Y: y - oy + hp},
		rl.Vector2{X: x - ox + wp, Y: y - oy + h},
		color,
	)
	rl.DrawTriangle(
		rl.Vector2{X: x - ox + wp, Y: y - oy},
		rl.Vector2{X: x - ox + wp, Y: y - oy + h},
		rl.Vector2{X: x - ox + wp*3, Y: y - oy + h},
		color,
	)
	rl.DrawTriangle(
		rl.Vector2{X: x - ox + wp*3, Y: y - oy},
		rl.Vector2{X: x - ox + wp, Y: y - oy},
		rl.Vector2{X: x - ox + wp*3, Y: y - oy + h},
		color,
	)
	rl.DrawTriangle(
		rl.Vector2{X: x - ox + wp*3, Y: y - oy},
		rl.Vector2{X: x - ox + wp*3, Y: y - oy + h},
		rl.Vector2{X: x - ox + w, Y: y - oy + hp},
		color,
	)
}
