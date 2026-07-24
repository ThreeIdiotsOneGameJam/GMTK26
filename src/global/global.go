package global

import (
	"sync/atomic"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

var WSState atomic.Value

var DebugEnabled = true

var MouseCursorState = rl.MouseCursorDefault
var MousePosition vec.Vec2
var ViewportSize = vec.Vec2i{X: 640, Y: 360}

var closeWindowRequested bool

func CloseWindow() {
	closeWindowRequested = true
}

func WindowShouldClose() bool {
	return rl.WindowShouldClose() || closeWindowRequested
}
