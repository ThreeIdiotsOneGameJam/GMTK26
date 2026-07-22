package world

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	v "github.com/threeidiotsonegamejam/gmtk26/src/math"
)

var Grid = [][]Tile{}
var Camera = rl.Camera2D{
	Zoom: 1.0,
}

var sqrt3 = float32(math.Sqrt(3.0))
var hexRadius = float32(64.0)

type Tile struct{
	Color rl.Color
}

func Init() {
	Grid = make([][]Tile, 32)

	for x := range len(Grid) {
		Grid[x] = make([]Tile, 12)
		for y := range len(Grid[x]) {
			Grid[x][y] = Tile{
				Color: rl.ColorLerp(rl.ColorLerp(rl.Green, rl.Lime, float32((x + y) % 2)), rl.Black, float32(x % 2) * 0.1),
			}
		}
	}
}

func Tick() {}

func Frame(delta float32) {
	
}

func Draw() {
	rl.BeginMode2D(Camera)

	rlMouse := rl.GetScreenToWorld2D(rl.GetMousePosition(), Camera)
	mp := v.Vec2{X: rlMouse.X, Y: rlMouse.Y}

	w := hexRadius * 2.0
	h := hexRadius * sqrt3
	for x := range len(Grid) {
		for y, tile := range Grid[x] {

			hex := PixelToHex(mp)

			tileColor := tile.Color
			if hex.X == int32(x) && hex.Y == int32(y) {
				tileColor = rl.ColorLerp(tileColor, rl.White, 0.2)
			}

			yOffset := float32(h/2.0) * float32(x % 2)
			DrawHexagon(float32(x) * w/4.0*3.0, float32(y) * h + yOffset, hexRadius, tileColor)
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
		rl.Vector2{ X: x-ox+wp, Y: y-oy },
		rl.Vector2{ X: x-ox,    Y: y-oy+hp },
		rl.Vector2{ X: x-ox+wp, Y: y-oy+h },
		color,
	)
	rl.DrawTriangle(
		rl.Vector2{ X: x-ox+wp,   Y: y-oy },
		rl.Vector2{ X: x-ox+wp,   Y: y-oy+h },
		rl.Vector2{ X: x-ox+wp*3, Y: y-oy+h },
		color,
	)
	rl.DrawTriangle(
		rl.Vector2{ X: x-ox+wp*3,   Y: y-oy },
		rl.Vector2{ X: x-ox+wp,     Y: y-oy },
		rl.Vector2{ X: x-ox+wp*3,   Y: y-oy+h },
		color,
	)
	rl.DrawTriangle(
	   	rl.Vector2{ X: x-ox+wp*3, Y: y-oy },
		rl.Vector2{ X: x-ox+wp*3, Y: y-oy+h },
		rl.Vector2{ X: x-ox+w,    Y: y-oy+hp },
		color,
	)
}
