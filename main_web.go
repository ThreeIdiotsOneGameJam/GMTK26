//go:build web

package main

import (
	"embed"

	"github.com/gen2brain/raylib-go/raylib"
)

//go:embed assets
var ASSETS embed.FS

func init() {
	rl.AddFileSystem(ASSETS)
}

func mainLoop() {
	rl.SetMainLoop(updateFunc)
	for !rl.WindowShouldClose() {
		updateFunc()
	}
}
