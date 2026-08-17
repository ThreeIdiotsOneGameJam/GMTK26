package render

import (
	"testing"

	"github.com/threeidiotsonegamejam/gmtk26/src/game"
)

func TestGetUnitRect(t *testing.T) {
	tests := []struct {
		unit game.UnitType
		x    float32
	}{
		{unit: game.UnitScout, x: 0},
		{unit: game.UnitPeasant, x: 96},
		{unit: game.UnitArcher, x: 192},
		{unit: game.UnitKnight, x: 288},
	}

	for _, test := range tests {
		rect := getUnitRect(test.unit)
		if rect.X != test.x || rect.Y != 0 ||
			rect.Width != unitTextureCellSize || rect.Height != unitTextureCellSize {
			t.Fatalf("%s: unexpected texture rect: %+v", test.unit, rect)
		}
	}

	if rect := getUnitRect(game.UnitUnknown); rect.Width != 0 || rect.Height != 0 {
		t.Fatalf("unknown unit should not have a texture rect: %+v", rect)
	}
}

func TestDrawUnitsSkipsNegativeFactionOwner(t *testing.T) {
	m := game.Map{
		Grid: [][]game.Cell{{
			{
				Units: []game.UnitData{{
					Type:  game.UnitScout,
					Owner: -1,
				}},
			},
		}},
	}

	renderer := WorldRenderer{}
	renderer.drawUnits(&m, []visibleTile{{
		hex: game.NewHex(0, 0),
	}})
}
