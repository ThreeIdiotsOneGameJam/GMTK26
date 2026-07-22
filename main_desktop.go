//go:build !web

package main

import "github.com/gen2brain/raylib-go/raylib"

func mainLoop() {
	for !rl.WindowShouldClose() {
		updateFunc()
	}
}
