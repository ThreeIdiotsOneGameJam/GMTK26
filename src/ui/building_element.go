package ui

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/rlvec"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

func Building(world *WorldElement) *BuildingElement {
	el := &BuildingElement{}
	el.BaseElement = NewBaseElement(el)

	el.World = world
	el.PosProvider = func(el *BuildingElement) vec.Vec2 {
		return el.AbsolutePos().Vec2()
	}

	return el
}

type BuildingElement struct {
	BaseElement[*BuildingElement]

	Type        game.BuildingType
	World       *WorldElement
	PosProvider func(el *BuildingElement) vec.Vec2

	tint    color.RGBA
	hex     game.Hex
	visible bool
}

func (el *BuildingElement) WithType(building game.BuildingType) *BuildingElement {
	el.Type = building
	return el
}

func (el *BuildingElement) WithPosProvider(provider func(el *BuildingElement) vec.Vec2) *BuildingElement {
	el.PosProvider = provider
	return el
}

func (el *BuildingElement) update(deltaNano int64) {
	pos := el.PosProvider(el)

	el.hex = el.World.Renderer.PixelToHex(rlvec.FromRL(rl.GetScreenToWorld2D(rlvec.ToRL(pos), el.World.Renderer.Camera)))

	if !el.World.Map.HexInsideBounds(el.hex) {
		el.visible = false
		return
	}

	el.visible = true

	if game.BuildingCanPlace(&el.World.Map, el.Type, el.hex) {
		el.tint = rl.Green
	} else {
		el.tint = rl.Red
	}

	el.tint.A = 123
}

func (el *BuildingElement) draw() {
	if !el.visible {
		return
	}
	el.World.Renderer.DrawBuilding(&el.World.Map, el.Type, el.hex, el.tint)
}
