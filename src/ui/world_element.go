package ui

import "github.com/threeidiotsonegamejam/gmtk26/src/world"

type WorldElement struct {
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
