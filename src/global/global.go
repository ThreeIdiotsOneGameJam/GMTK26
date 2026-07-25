package global

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

var DebugAvailable = false
var DebugEnabled = false

var MouseCursorState = rl.MouseCursorDefault
var MousePosition vec.Vec2

// UIBlocksWorldInput is set by interactive UI while the cursor is over it
// (or while it is actively handling a drag/click) so world clicks do not pass through.
var UIBlocksWorldInput bool
var ViewportSize = vec.Vec2i{X: 640, Y: 360}

var closeWindowRequested bool

func CloseWindow() {
	closeWindowRequested = true
}

func WindowShouldClose() bool {
	return rl.WindowShouldClose() || closeWindowRequested
}
