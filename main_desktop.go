//go:build !web

package main

import (
	"github.com/threeidiotsonegamejam/gmtk26/src/global"
)

func mainLoop() {
	for !global.WindowShouldClose() {
		updateFunc()
	}
}
