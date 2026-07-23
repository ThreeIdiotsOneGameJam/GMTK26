package ui

import "github.com/threeidiotsonegamejam/gmtk26/src/world"

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
	w.World.Draw()
}

func (w *WorldElement) draw() {
	w.World.Draw()
}
