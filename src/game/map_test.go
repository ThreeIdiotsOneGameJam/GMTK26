package game

import (
	"testing"

	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

func TestHexInsideBoundsRequiresActualGridCell(t *testing.T) {
	m := Map{
		GridSize: vec.Vec2i{X: 2, Y: 2},
		Grid: [][]Cell{
			{{Tile: TilePlains}},
		},
	}

	if m.HexInsideBounds(NewHex(1, 1)) {
		t.Fatal("missing grid cell reported inside bounds")
	}
	if !m.HexInsideBounds(NewHex(0, 0)) {
		t.Fatal("existing land cell reported outside bounds")
	}
}

func TestNilMapHasNoHexInsideBounds(t *testing.T) {
	var m *Map
	if m.HexInsideBounds(NewHex(0, 0)) {
		t.Fatal("nil map reported a hex inside bounds")
	}
}
