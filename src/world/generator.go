package world

import (
	// "image/color"
	"math"
	"math/rand"

	"github.com/aquilax/go-perlin"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

type Generator struct {
	Seed   int64
	Random *rand.Rand
}

func (g *Generator) Generate(size vec.Vec2i) [][]Cell {
	noise := perlin.NewPerlin(2, 3, 2, g.Seed)

	sampleNoise := func(x, y float64) float64 {
		sx := x * 10.0
		sy := y * 10.0
		return noise.Noise2D(sx, sy)*0.5 + 0.5
	}

	grid := make([][]Cell, size.X)
	center := size.Vec2().Mul(vec.Vec2{X: 0.5, Y: 0.5})

	for x := range size.X {
		grid[x] = make([]Cell, size.Y)
		for y := range size.Y {
			var tile Tile = &WaterTile{}

			fx, fy := float64(x), float64(y)
			vx, vy := fx/float64(size.X), fy/float64(size.Y)

			noise := sampleNoise(vx, vy)
			distance := float64(1.0 - (center.Distance(vec.Vec2i{X: x, Y: y}.Vec2())/(size.Vec2().Y*0.5) - 0.2))
			height := distance * (noise*0.5 + 0.2) * math.Min(sampleNoise(vx, vy)*2.0, 1.0)
			if height < 0.0 {
				height = 0.0
			}
			// value256 := uint8(height * 255)

			if x == 0 || x == size.X-1 || y == 0 || y == size.Y-1 || distance < 0.2 {
				tile = &VoidTile{}
				goto finish_tile
			}

			if height > 0.25 {
				tile = &GrassTile{}
			} else if height > 0.2 {
				tile = &UnkownTile{}
			} else {
				tile = &WaterTile{}
			}

			// tile = &ColorTile{
			// 	Color: color.RGBA{R: value256, G: value256, B: value256, A: 255},
			// }

		finish_tile:
			grid[x][y] = Cell{Tile: tile}
		}
	}

	return grid
}
