package game

import (
	"math"
	"math/rand"

	"github.com/aquilax/go-perlin"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

func generateMap(size vec.Vec2i, seed int64) [][]Cell {
	noise := perlin.NewPerlin(2, 3, 2, seed)

	sampleNoise := func(x, y float64) float64 {
		return noise.Noise2D(x*10.0, y*10.0)*0.5 + 0.5
	}

	grid := make([][]Cell, size.X)
	centerX := float64(size.X) * 0.5
	centerY := float64(size.Y) * 0.5

	for x := range size.X {
		grid[x] = make([]Cell, size.Y)
		for y := range size.Y {
			tile := TilePlains
			fx, fy := float64(x), float64(y)
			vx, vy := fx/float64(size.X), fy/float64(size.Y)
			distanceFromCenter := math.Hypot(fx-centerX, fy-centerY)
			distance := 1.0 - (distanceFromCenter/(float64(size.Y)*0.5) - 0.2)
			height := distance * (sampleNoise(vx, vy)*0.5 + 0.2) * math.Min(sampleNoise(vx, vy)*2.0, 1.0)
			if height < 0.0 {
				height = 0.0
			}

			const moistureNoiseOffset = 17.6321
			moisture := sampleNoise(vx+moistureNoiseOffset, vy+moistureNoiseOffset)

			const temperatureNoiseOffset = 29.13329
			temperature := sampleNoise(vx+temperatureNoiseOffset, vy+temperatureNoiseOffset)

			if moisture > 0.6 {
				tile = TileForest
			} else if moisture < 0.45 {
				tile = TileRock
			}

			if temperature > 0.6 {
				tile = TileDesert
				if moisture > 0.6 {
					tile = TileJungle
				}
			}

			if height < 0.25 {
				tile = TileWater
			}

			if x == 0 || x == size.X-1 || y == 0 || y == size.Y-1 || distance-sampleNoise(vx*0.5-2.32, vy*0.5-2.32)*0.3 < 0.2 {
				tile = TileVoid
			}

			grid[x][y] = Cell{Tile: tile, Owner: -1}
		}
	}

	return grid
}

func spreadResources(m *Map, seed int64) {
	rock := make([]Hex, 0)
	for x := range m.Grid {
		for y := range m.Grid[x] {
			if m.Grid[x][y].Tile == TileRock {
				rock = append(rock, NewHex(int32(x), int32(y)))
			}
		}
	}

	r := rand.New(rand.NewSource(seed))
	r.Shuffle(len(rock), func(i, j int) {
		rock[i], rock[j] = rock[j], rock[i]
	})

	setTiles := func(tile TileType, count int) {
		count = min(count, len(rock))
		for _, pos := range rock[:count] {
			m.Grid[pos.X][pos.Y].Tile = tile
		}
		rock = rock[count:]
	}

	rockCount := len(rock)
	setTiles(TileIron, rockCount/10)
	setTiles(TileCoal, rockCount/10)
	setTiles(TileGold, rockCount/20)
}
