package world

import (
	"math"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/mathutil/vec"
	"github.com/threeidiotsonegamejam/gmtk26/src/util"
)

type Generator struct {
	Seed   int64
	Random *rand.Rand
}

func (g Generator) Generate(size vec.Vec2i) [][]Cell {
	g.Random = rand.New(rand.NewSource(g.Seed))
	perlin := rl.GenImagePerlinNoise(512, 512, int(g.Random.Float32()), 0, 0.7)
	defer rl.UnloadImage(perlin)

	sampleNoise := func(x, y float32) float32 {
		x1 := float32(math.Floor(float64(x * float32(perlin.Width))))
		x2 := x1 + 1.0
		xt := x - x1
		if x2 >= float32(perlin.Width) {
			x2 = 1.0
		}

		y1 := float32(math.Floor(float64(y * float32(perlin.Height))))
		y2 := y1 + 1.0
		yt := y - y1
		if y2 >= float32(perlin.Height) {
			y2 = 1.0
		}

		t := (xt + yt) / 2.0

		a := float32(util.SampleImageRepeat(*perlin, int32(x1), int32(y1)).R) / 255.0
		b := float32(util.SampleImageRepeat(*perlin, int32(x2), int32(y2)).R) / 255.0

		return a*(1.0-t) + b*t
	}

	grid := make([][]Cell, size.X)
	center := size.Vec2().Mul(vec.Vec2{X: 0.5, Y: 0.5})

	for x := range size.X {
		grid[x] = make([]Cell, size.Y)
		for y := range size.Y {
			var tile Tile = &WaterTile{}

			noise := sampleNoise(float32(x)/float32(size.X), float32(y)/float32(size.Y))
			distance := 1.0 - center.Distance(vec.Vec2i{X: x, Y: y}.Vec2())/(size.Vec2().Y*0.5)
			height := distance * (noise*0.5 + 0.2)
			if height < 0.0 {
				height = 0.0
			}

			if x == 0 || x == size.X-1 || y == 0 || y == size.Y-1 || distance < 0.01 {
				tile = &VoidTile{}
				goto finish_tile
			}

			if height > 0.6 {
				tile = &GrassTile{}
			} else if height > 0.4 {
				tile = &UnkownTile{}
			} else {
				tile = &WaterTile{}
			}

		finish_tile:
			grid[x][y] = Cell{Tile: tile}
		}
	}

	return grid
}
