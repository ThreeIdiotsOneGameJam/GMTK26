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
	Map      game.Map
	Renderer render.WorldRenderer
}

func (el *GameWorldElement) prepare() {
	if el.Map.Grid == nil {
		el.Map.Generate()
	}
	el.Renderer.Init(&el.Map)
}

func (el *GameWorldElement) update(deltaNano int64) {
	el.Renderer.Update(&el.Map, time.Duration(deltaNano))
	audio.AmbienceVolumeMulti = el.Renderer.TargetZoom*0.5 + 0.5
}

func (el *GameWorldElement) draw() {
	el.Renderer.Draw(&el.Map)
}
