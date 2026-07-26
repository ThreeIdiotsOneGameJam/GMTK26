package ui

import (
	"fmt"

	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/render"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

var gameSelectionMenuSize = vec.Vec2i{X: 236, Y: 128}

const (
	gameSelectionMenuMargin int32 = 12
	gameSelectionMenuOffset int32 = 18
)

// GameSelectionMenu is a world-anchored UI popover for tiles that contain
// both a unit and a building. It is composed from normal Panel, Text, and
// Button elements so it participates in the shared UI input and styling.
func GameSelectionMenu() *GameSelectionMenuElement {
	el := &GameSelectionMenuElement{}
	el.BaseElement = NewBaseElement(el)

	el.unitButton = Button().
		WithTextSize(18).
		WithPadding(6).
		WithOutlineWidth(2).
		WithSize(vec.Vec2i{X: 212, Y: 34}).
		WithRelativePos(vec.Vec2i{X: 12, Y: 40}).
		WithBackgroundColors(ColorSet{
			Default:  ptrColor(PaletteSurface),
			Hover:    ptrColor(PaletteIndigoDim),
			Click:    ptrColor(PaletteIndigoPress),
			Disabled: ptrColor(PaletteSurface),
		}).
		WithEnabledDynamic(func(_ *ButtonElement) bool {
			return el.unitAvailable
		}).
		WithClick(func() {
			if el.world != nil {
				el.world.Renderer.ChooseSelectionMenuOption(
					&el.world.Map,
					render.SelectionUnit,
				)
			}
		})

	el.buildingButton = Button().
		WithTextSize(18).
		WithPadding(6).
		WithOutlineWidth(2).
		WithSize(vec.Vec2i{X: 212, Y: 34}).
		WithRelativePos(vec.Vec2i{X: 12, Y: 82}).
		WithBackgroundColors(ColorSet{
			Default:  ptrColor(PaletteSurface),
			Hover:    ptrColor(PaletteIndigoDim),
			Click:    ptrColor(PaletteIndigoPress),
			Disabled: ptrColor(PaletteSurface),
		}).
		WithEnabledDynamic(func(_ *ButtonElement) bool {
			return el.buildingAvailable
		}).
		WithClick(func() {
			if el.world != nil {
				el.world.Renderer.ChooseSelectionMenuOption(
					&el.world.Map,
					render.SelectionBuilding,
				)
			}
		})

	return el.
		WithSize(gameSelectionMenuSize).
		WithRelativePosDynamic(func(el *GameSelectionMenuElement) vec.Vec2i {
			return el.position()
		}).
		WithVisibleDynamic(func(el *GameSelectionMenuElement) bool {
			if el.world == nil {
				return false
			}
			_, visible := el.world.Renderer.SelectionMenuHex()
			return visible
		}).
		AddChild(
			Panel().
				WithSize(gameSelectionMenuSize).
				WithRoundness(0).
				WithWorldInputBlocking(true),
		).
		AddChild(
			Text().
				WithText("SELECT ON TILE").
				WithTextSize(16).
				WithTextColor(PaletteTextSecondary).
				WithRelativePos(vec.Vec2i{X: 12, Y: 12}),
		).
		AddChild(el.unitButton).
		AddChild(el.buildingButton)
}

func (el *GameSelectionMenuElement) WithWorld(world *GameWorldElement) *GameSelectionMenuElement {
	el.world = world
	return el
}

type GameSelectionMenuElement struct {
	BaseElement[*GameSelectionMenuElement]

	world             *GameWorldElement
	unitButton        *ButtonElement
	buildingButton    *ButtonElement
	unitAvailable     bool
	buildingAvailable bool
}

func (el *GameSelectionMenuElement) update(deltaNano int64) {
	if el.world == nil {
		return
	}

	hex, visible := el.world.Renderer.SelectionMenuHex()
	if !visible {
		return
	}
	cell := el.world.Map.GetCell(hex)
	if cell == nil ||
		(cell.Unit == game.UnitUnknown && cell.Building == game.BuildingUnknown) {
		el.world.Renderer.DismissSelectionMenu()
		return
	}

	el.unitAvailable = cell.Unit != game.UnitUnknown
	el.buildingAvailable = cell.Building != game.BuildingUnknown
	el.unitButton.Text = fmt.Sprintf("%s unit", cell.Unit)
	el.buildingButton.Text = fmt.Sprintf("%s building", cell.Building)
}

func (el *GameSelectionMenuElement) draw() {}

func (el *GameSelectionMenuElement) position() vec.Vec2i {
	if el.world == nil {
		return vec.Vec2i{}
	}
	hex, visible := el.world.Renderer.SelectionMenuHex()
	if !visible {
		return vec.Vec2i{}
	}

	screenSize := vec.Vec2i{}
	if root := el.RootElement(); root != nil {
		screenSize = root.Size()
	}
	return selectionMenuPosition(
		el.world.Renderer.HexScreenPosition(hex),
		screenSize,
		el.Size(),
	)
}

func selectionMenuPosition(
	tileCenter vec.Vec2,
	screenSize vec.Vec2i,
	menuSize vec.Vec2i,
) vec.Vec2i {
	anchorPosition := tileCenter.RoundToInt()
	x := anchorPosition.X + gameSelectionMenuOffset
	y := anchorPosition.Y - gameSelectionMenuOffset

	// Prefer the tile's right side, then flip to its left before clamping so
	// the menu does not cover the object the player is choosing between.
	if x+menuSize.X+gameSelectionMenuMargin > screenSize.X {
		x = anchorPosition.X - menuSize.X - gameSelectionMenuOffset
	}

	maxX := max(gameSelectionMenuMargin, screenSize.X-menuSize.X-gameSelectionMenuMargin)
	maxY := max(gameSelectionMenuMargin, screenSize.Y-menuSize.Y-gameSelectionMenuMargin)
	x = max(gameSelectionMenuMargin, min(x, maxX))
	y = max(gameSelectionMenuMargin, min(y, maxY))
	return vec.Vec2i{X: x, Y: y}
}
