package world

import (
	"math"

	"github.com/aquilax/go-perlin"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

func Generate(size vec.Vec2i, seed int64) [][]Cell {
	noise := perlin.NewPerlin(2, 3, 2, seed)

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
			tile := PlainsTile

			fx, fy := float64(x), float64(y)
			vx, vy := fx/float64(size.X), fy/float64(size.Y)

			distance := float64(1.0 - (center.Distance(vec.Vec2i{X: x, Y: y}.Vec2())/(size.Vec2().Y*0.5) - 0.2))
			height := distance * (sampleNoise(vx, vy)*0.5 + 0.2) * math.Min(sampleNoise(vx, vy)*2.0, 1.0)
			if height < 0.0 {
				height = 0.0
			}

			const moistureNoiseOffset = 17.6321
			moisture := sampleNoise(vx+moistureNoiseOffset, vy+moistureNoiseOffset)

			const temperatureNoiseOffset = 29.13329
			temperature := sampleNoise(vx+temperatureNoiseOffset, vy+temperatureNoiseOffset)

			if moisture > 0.6 {
				tile = ForestTile
			} else if moisture < 0.45 {
				tile = RockTile
			}

			if temperature > 0.6 {
				tile = DesertTile
				if moisture > 0.6 {
					tile = JungleTile
				}
			}

			if height < 0.25 {
				tile = WaterTile
			}

			if x == 0 || x == size.X-1 || y == 0 || y == size.Y-1 || distance-sampleNoise(vx*0.5-2.32, vy*0.5-2.32)*0.3 < 0.2 {
				tile = VoidTile
			}

			grid[x][y] = Cell{Tile: tile}
		}
	}

	return grid
}
