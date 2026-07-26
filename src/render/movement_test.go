package render

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

func TestMovementOrderRouteIncludesPersistentTurnStops(t *testing.T) {
	grid := make([][]game.Cell, 1)
	grid[0] = make([]game.Cell, 6)
	for y := range grid[0] {
		grid[0][y] = game.Cell{Tile: game.TilePlains, Owner: -1}
	}
	m := game.Map{
		Grid:     grid,
		GridSize: vec.Vec2i{X: 1, Y: 6},
	}
	start := game.NewHex(0, 0)
	destination := game.NewHex(0, 5)
	m.GetCell(start).Unit = game.UnitPeasant
	m.GetCell(start).UnitOwner = 0

	path, stops, ok := movementOrderRoute(&m, 0, game.MovementOrder{
		Current:     start,
		Destination: destination,
	})
	if !ok {
		t.Fatal("expected queued movement route to remain drawable")
	}
	if len(path) != 6 {
		t.Fatalf("path length = %d, want 6", len(path))
	}

	want := []game.Hex{
		game.NewHex(0, 2),
		game.NewHex(0, 4),
		game.NewHex(0, 5),
	}
	if len(stops) != len(want) {
		t.Fatalf("stops = %v, want %v", stops, want)
	}
	for i := range want {
		if stops[i] != want[i] {
			t.Fatalf("stops = %v, want %v", stops, want)
		}
	}
}

func TestZoomSafeSizeOnlyClampsAtDeepZoom(t *testing.T) {
	r := WorldRenderer{}
	r.Camera.Zoom = 0.75
	if got := r.zoomSafeSize(3, 0.9); got != 3 {
		t.Fatalf("normal zoom width = %f, want original world width 3", got)
	}

	r.Camera.Zoom = 0.08
	if got := r.zoomSafeSize(3, 0.9) * r.Camera.Zoom; got < 0.899 || got > 0.901 {
		t.Fatalf("deep zoom screen width = %f, want 0.9", got)
	}
}

func TestRightClickOnQueuedRouteCancelsIt(t *testing.T) {
	grid := make([][]game.Cell, 1)
	grid[0] = []game.Cell{
		{Tile: game.TilePlains, Owner: -1, Unit: game.UnitScout, UnitOwner: 0},
		{Tile: game.TilePlains, Owner: -1},
		{Tile: game.TilePlains, Owner: -1},
	}
	m := game.Map{
		Grid:     grid,
		GridSize: vec.Vec2i{X: 1, Y: 3},
	}
	start := game.NewHex(0, 0)
	destination := game.NewHex(0, 2)
	cancelledFrom := game.NewHex(-1, -1)
	r := WorldRenderer{
		Camera:       rl.Camera2D{Zoom: 0.25},
		HexSize:      vec.Vec2{X: 48, Y: 48},
		LocalFaction: 0,
		Orders: []game.MovementOrder{{
			Current:     start,
			Destination: destination,
		}},
		OnCancelMovement: func(from game.Hex) bool {
			cancelledFrom = from
			return true
		},
	}
	first := r.HexToPixel(start.Vec2i)
	second := r.HexToPixel(game.NewHex(0, 1).Vec2i)
	onRoute := first.Lerp(second, 0.5)

	if r.cancelQueuedMovementAt(&m, onRoute, false) {
		t.Fatal("right drag cancelled queued route")
	}
	if !r.cancelQueuedMovementAt(&m, onRoute, true) {
		t.Fatal("stationary right click on route was not handled")
	}
	if cancelledFrom != start {
		t.Fatalf("cancelled source = %v, want %v", cancelledFrom, start)
	}
	if len(r.Orders) != 0 {
		t.Fatal("cancelled route remained visible")
	}
}
