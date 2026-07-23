package ui

import (
	"time"

	"github.com/threeidiotsonegamejam/gmtk26/src/world"
)

func World() *WorldElement {
	el := &WorldElement{}
	el.BaseElement = NewBaseElement(el)

	return el
}

type WorldElement struct {
	BaseElement[*WorldElement]
	World world.World
}

func (w *WorldElement) update(deltaNano int64) {
	if !w.World.HasInit {
		w.World.Init()
	}
	w.World.Update(float32(deltaNano / int64(time.Nanosecond)))
}

func (w *WorldElement) draw() {
	w.World.Draw()
}
