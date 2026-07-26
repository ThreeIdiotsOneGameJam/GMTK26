package ui

import (
	"testing"

	"github.com/threeidiotsonegamejam/gmtk26/src/global"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

func TestSelectionMenuPositionFollowsTile(t *testing.T) {
	screen := vec.Vec2i{X: 1280, Y: 720}
	size := gameSelectionMenuSize

	first := selectionMenuPosition(vec.Vec2{X: 400, Y: 300}, screen, size)
	second := selectionMenuPosition(vec.Vec2{X: 455, Y: 275}, screen, size)

	if second.X-first.X != 55 || second.Y-first.Y != -25 {
		t.Fatalf(
			"menu movement = (%d, %d), want (55, -25)",
			second.X-first.X,
			second.Y-first.Y,
		)
	}
}

func TestSelectionMenuPositionFlipsAndClampsAtWindowEdges(t *testing.T) {
	screen := vec.Vec2i{X: 800, Y: 600}
	size := gameSelectionMenuSize

	right := selectionMenuPosition(vec.Vec2{X: 780, Y: 300}, screen, size)
	if right.X >= 780 {
		t.Fatalf("right-edge menu x = %d, expected it left of the tile", right.X)
	}

	top := selectionMenuPosition(vec.Vec2{X: 300, Y: -100}, screen, size)
	if top.Y != gameSelectionMenuMargin {
		t.Fatalf("top-edge menu y = %d, want %d", top.Y, gameSelectionMenuMargin)
	}

	bottom := selectionMenuPosition(vec.Vec2{X: 300, Y: 900}, screen, size)
	wantBottom := screen.Y - size.Y - gameSelectionMenuMargin
	if bottom.Y != wantBottom {
		t.Fatalf("bottom-edge menu y = %d, want %d", bottom.Y, wantBottom)
	}
}

func TestSelectionMenuPanelPaddingBlocksWorldInput(t *testing.T) {
	previousMouse := global.MousePosition
	previousBlocks := global.UIBlocksWorldInput
	previousModal := global.UIModalBlocksInput
	t.Cleanup(func() {
		global.MousePosition = previousMouse
		global.UIBlocksWorldInput = previousBlocks
		global.UIModalBlocksInput = previousModal
	})

	panel := Panel().
		WithRelativePos(vec.Vec2i{X: 100, Y: 80}).
		WithSize(gameSelectionMenuSize).
		WithWorldInputBlocking(true)
	global.MousePosition = vec.Vec2{X: 101, Y: 81}
	global.UIBlocksWorldInput = false
	global.UIModalBlocksInput = false

	panel.update(0)

	if !global.UIBlocksWorldInput {
		t.Fatal("pointer over menu padding did not block the world")
	}
}

func TestSelectionMenuUsesSquareStandardPanelBorder(t *testing.T) {
	menu := GameSelectionMenu()
	panel, ok := menu.Children[0].(*PanelElement)
	if !ok {
		t.Fatalf("first menu child = %T, want *PanelElement", menu.Children[0])
	}
	if panel.Roundness != 0 {
		t.Fatalf("menu panel roundness = %f, want 0", panel.Roundness)
	}
	if panel.OutlineColor != PaletteBorder {
		t.Fatalf("menu outline = %#v, want standard border %#v", panel.OutlineColor, PaletteBorder)
	}
}
