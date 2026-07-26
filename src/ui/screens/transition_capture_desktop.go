//go:build !web

package screens

import rl "github.com/gen2brain/raylib-go/raylib"

func flushTransitionSourceDraws() {
	rl.DrawRenderBatchActive()
}
