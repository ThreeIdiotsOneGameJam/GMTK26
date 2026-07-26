//go:build web

package render

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

// hexBatch groups hexagons by color and submits each group in one raylib call.
// That is important on web, where every individual rlgl vertex call crosses
// the Go WebAssembly -> JavaScript -> raylib WebAssembly boundary.
type hexBatch struct {
	strips map[color.RGBA][]rl.Vector2
}

func newHexBatch() *hexBatch {
	return &hexBatch{strips: make(map[color.RGBA][]rl.Vector2)}
}

func (b *hexBatch) Add(x, y float32, size vec.Vec2, tint color.RGBA) {
	if tint.A == 0 {
		return
	}

	width := size.X * 2.0
	height := size.Y * sqrt3
	wp := width / 4.0
	hp := height / 2.0
	ox := width / 2.0
	oy := hp

	a := rl.Vector2{X: x - ox + wp, Y: y - oy}
	bb := rl.Vector2{X: x - ox, Y: y - oy + hp}
	c := rl.Vector2{X: x - ox + wp, Y: y - oy + height}
	d := rl.Vector2{X: x - ox + wp*3.0, Y: y - oy + height}
	e := rl.Vector2{X: x - ox + width, Y: y - oy + hp}
	f := rl.Vector2{X: x - ox + wp*3.0, Y: y - oy}
	center := rl.Vector2{X: x, Y: y}

	// This strip draws the six triangles of a hexagonal fan. Repeated center
	// vertices make the intervening strip triangles degenerate.
	hex := [...]rl.Vector2{a, bb, center, c, center, d, center, e, center, f, center, a}
	points := b.strips[tint]
	if len(points) == 0 {
		b.strips[tint] = append(points, hex[:]...)
		return
	}

	// Join separate hex strips with degenerate triangles. Each appended hex
	// contributes an even number of vertices, preserving strip winding.
	last := points[len(points)-1]
	points = append(points, last, hex[0], hex[0])
	points = append(points, hex[1:]...)
	b.strips[tint] = points
}

func (b *hexBatch) Draw() {
	for tint, points := range b.strips {
		rl.DrawTriangleStrip(points, tint)
	}
}
