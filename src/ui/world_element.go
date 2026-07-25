package ui

import (
	"time"

	"github.com/threeidiotsonegamejam/gmtk26/src/audio"
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/render"
)

func World() *WorldElement {
	el := &WorldElement{}
	el.BaseElement = NewBaseElement(el)

	return el
}

type WorldElement struct {
	BaseElement[*WorldElement]
	Map      game.Map
	Renderer render.WorldRenderer
}

func (el *WorldElement) prepare() {
	if el.Map.Grid == nil {
		el.Map.Generate()
	}
	el.Renderer.Init(&el.Map)
}

func (el *WorldElement) update(deltaNano int64) {
	el.Renderer.Update(&el.Map, time.Duration(deltaNano))
	audio.AmbienceVolumeMulti = el.Renderer.TargetZoom*0.5 + 0.5
}

func (el *WorldElement) draw() {
	el.Renderer.Draw(&el.Map)
}
