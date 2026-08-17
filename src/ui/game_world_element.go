package ui

import (
	"time"

	"github.com/threeidiotsonegamejam/gmtk26/src/audio"
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/render"
)

func GameWorld() *GameWorldElement {
	el := &GameWorldElement{}
	el.BaseElement = NewBaseElement(el)

	return el
}

type GameWorldElement struct {
	BaseElement[*GameWorldElement]
	Map          game.Map
	Renderer     render.WorldRenderer
	pendingFocus *game.Hex
}

func (el *GameWorldElement) prepare() {
	el.Renderer.Init(&el.Map)
	if el.pendingFocus != nil {
		el.Renderer.FocusOnHex(*el.pendingFocus)
		el.pendingFocus = nil
	}
}

func (el *GameWorldElement) update(deltaNano int64) {
	el.Renderer.Update(&el.Map, time.Duration(deltaNano))
	audio.AmbienceVolumeMulti = el.Renderer.TargetZoom*0.5 + 0.5
}

func (el *GameWorldElement) draw() {
	el.Renderer.Draw(&el.Map)
}

// FocusOnHex defers camera focus until prepare has initialized the renderer.
// This matters when the world stays hidden throughout the lobby.
func (el *GameWorldElement) FocusOnHex(hex game.Hex) {
	el.pendingFocus = &hex
}
