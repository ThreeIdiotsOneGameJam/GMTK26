package world

import (
	"image/color"
)

type Tile interface {
	Color() color.RGBA
	Tick()
	Update(delta float32)
	Draw()
}

type VoidTile struct{}

func (t *VoidTile) Color() color.RGBA {
	return color.RGBA{R: 0, G: 0, B: 0, A: 0}
}
func (t *VoidTile) Tick()                {}
func (t *VoidTile) Update(delta float32) {}
func (t *VoidTile) Draw()                {}

type UnkownTile struct{}

func (t *UnkownTile) Color() color.RGBA {
	return color.RGBA{R: 211, G: 191, B: 139, A: 255}
}
func (t *UnkownTile) Tick()                {}
func (t *UnkownTile) Update(delta float32) {}
func (t *UnkownTile) Draw()                {}

// These tiles are temporary
type WaterTile struct{}

func (t *WaterTile) Color() color.RGBA {
	return color.RGBA{R: 50, G: 70, B: 200, A: 255}
}
func (t *WaterTile) Tick()                {}
func (t *WaterTile) Update(delta float32) {}
func (t *WaterTile) Draw()                {}

type GrassTile struct{}

func (t *GrassTile) Color() color.RGBA {
	return color.RGBA{R: 63, G: 200, B: 140, A: 255}
}
func (t *GrassTile) Tick()                {}
func (t *GrassTile) Update(delta float32) {}
func (t *GrassTile) Draw()                {}
