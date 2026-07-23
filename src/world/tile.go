package world

import (
	"image/color"

	v "github.com/threeidiotsonegamejam/gmtk26/src/mathutil/vec"
)

type TileData struct {
	Type  string
	Color color.RGBA
}

type Tile interface {
	Data() TileData
	Tick()
	Update(w World, delta float32, pos v.Vec2)
	Draw(w World, pos v.Vec2)
}

type VoidTile struct{}

func (t *VoidTile) Data() TileData {
	return TileData{
		Type:  "void",
		Color: color.RGBA{R: 0, G: 0, B: 0, A: 0},
	}
}
func (t *VoidTile) Tick()                                     {}
func (t *VoidTile) Update(w World, delta float32, pos v.Vec2) {}
func (t *VoidTile) Draw(w World, pos v.Vec2)                  {}

type UnkownTile struct{}

func (t *UnkownTile) Data() TileData {
	return TileData{
		Type:  "unkown",
		Color: color.RGBA{R: 211, G: 191, B: 139, A: 255},
	}
}
func (t *UnkownTile) Tick()                                     {}
func (t *UnkownTile) Update(w World, delta float32, pos v.Vec2) {}
func (t *UnkownTile) Draw(w World, pos v.Vec2)                  {}

// These tiles are temporary
type WaterTile struct{}

func (t *WaterTile) Data() TileData {
	return TileData{
		Type:  "water",
		Color: color.RGBA{R: 50, G: 70, B: 200, A: 255},
	}
}
func (t *WaterTile) Tick()                                     {}
func (t *WaterTile) Update(w World, delta float32, pos v.Vec2) {}
func (t *WaterTile) Draw(w World, pos v.Vec2)                  {}

type GrassTile struct{}

func (t *GrassTile) Data() TileData {
	return TileData{
		Type:  "grass",
		Color: color.RGBA{R: 63, G: 200, B: 140, A: 255},
	}
}
func (t *GrassTile) Tick()                                     {}
func (t *GrassTile) Update(w World, delta float32, pos v.Vec2) {}
func (t *GrassTile) Draw(w World, pos v.Vec2)                  {}
