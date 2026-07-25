package render

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/util"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

type visibleTile struct {
	hex      game.Hex
	position vec.Vec2
	tile     game.TileType
}

func tileColor(tile game.TileType) color.RGBA {
	switch tile {
	case game.TileVoid:
		return color.RGBA{}
	case game.TileWater:
		return color.RGBA{R: 50, G: 70, B: 200, A: 255}
	case game.TilePlains:
		return color.RGBA{R: 50, G: 200, B: 80, A: 255}
	case game.TileForest:
		return color.RGBA{R: 50, G: 150, B: 100, A: 255}
	case game.TileDesert:
		return color.RGBA{R: 196, G: 190, B: 100, A: 255}
	case game.TileJungle:
		return color.RGBA{R: 50, G: 120, B: 20, A: 255}
	case game.TileRock, game.TileIron, game.TileCoal, game.TileGold:
		return color.RGBA{R: 150, G: 150, B: 150, A: 255}
	default:
		return rl.Magenta
	}
}

func (r *WorldRenderer) drawTileDetails(m *game.Map, visible []visibleTile) {
	r.drawEdges(m, visible, game.TileVoid, tileColor(game.TileVoid))
	r.drawEdges(m, visible, game.TileWater, *util.ColorSub(tileColor(game.TileWater), 50))

	rl.Begin(rl.Triangles)
	for _, tile := range visible {
		size := r.HexSize.Sub(vec.Vec2{X: 16.0, Y: 16.0})
		switch tile.tile {
		case game.TileIron:
			drawHexagonBuffered(tile.position.X, tile.position.Y, size, rl.ColorLerp(rl.Brown, rl.White, 0.5))
		case game.TileCoal:
			drawHexagonBuffered(tile.position.X, tile.position.Y, size, rl.Black)
		case game.TileGold:
			drawHexagonBuffered(tile.position.X, tile.position.Y, size, rl.Gold)
		}
	}
	rl.End()
}

func (r *WorldRenderer) drawEdges(m *game.Map, visible []visibleTile, tileType game.TileType, edgeColor color.RGBA) {
	if !rl.IsShaderValid(r.voidShader) {
		return
	}

	hasTiles := false
	for _, tile := range visible {
		if tile.tile == tileType {
			hasTiles = true
			break
		}
	}
	if !hasTiles {
		return
	}

	rl.BeginShaderMode(r.voidShader)
	rl.Begin(rl.Triangles)
	for _, tile := range visible {
		if tile.tile == tileType {
			r.drawEdge(m, tile.hex, tile.position, tileType, edgeColor)
		}
	}
	rl.End()
	rl.EndShaderMode()
}

func (r *WorldRenderer) drawEdge(m *game.Map, tile game.Hex, position vec.Vec2, tileType game.TileType, edgeColor color.RGBA) {
	neighbors := m.GetNeighbors(tile)
	isEdge := func(cell *game.Cell) bool {
		return cell != nil && cell.Tile != tileType
	}

	if !isEdge(neighbors.N) && !isEdge(neighbors.NE) && !isEdge(neighbors.NW) &&
		!isEdge(neighbors.S) && !isEdge(neighbors.SE) && !isEdge(neighbors.SW) {
		return
	}

	x, y := position.X, position.Y
	width := r.HexSize.X * 2.0
	height := r.HexSize.Y * sqrt3
	wp := width / 4.0
	hp := height / 2.0
	ox := width / 2.0
	oy := hp
	a := vec.Vec2{X: x - ox + wp, Y: y - oy}
	b := vec.Vec2{X: x - ox, Y: y - oy + hp}
	c := vec.Vec2{X: x - ox + wp, Y: y - oy + height}
	d := vec.Vec2{X: x - ox + wp*3, Y: y - oy + height}
	e := vec.Vec2{X: x - ox + width, Y: y - oy + hp}
	f := vec.Vec2{X: x - ox + wp*3, Y: y - oy}
	center := vec.Vec2{X: x, Y: y}

	drawSection := func(v1, v2 vec.Vec2, border1, border2, edge bool) {
		if (border1 || border2) && !edge {
			mid := v1.Lerp(v2, 0.5)
			setEdgeNormal(border1)
			rl.Vertex2f(v1.X, v1.Y)
			setEdgeNormal(false)
			rl.Vertex2f(mid.X, mid.Y)
			rl.Vertex2f(center.X, center.Y)
			rl.Vertex2f(mid.X, mid.Y)
			setEdgeNormal(border2)
			rl.Vertex2f(v2.X, v2.Y)
			setEdgeNormal(false)
			rl.Vertex2f(center.X, center.Y)
			return
		}

		setEdgeNormal(edge)
		rl.Vertex2f(v1.X, v1.Y)
		rl.Vertex2f(v2.X, v2.Y)
		setEdgeNormal(false)
		rl.Vertex2f(center.X, center.Y)
	}

	rl.Color4ub(edgeColor.R, edgeColor.G, edgeColor.B, edgeColor.A)
	rl.TexCoord2f(position.X, position.Y)
	drawSection(a, b, isEdge(neighbors.N), isEdge(neighbors.SW), isEdge(neighbors.NW))
	drawSection(b, c, isEdge(neighbors.NW), isEdge(neighbors.S), isEdge(neighbors.SW))
	drawSection(c, d, isEdge(neighbors.SW), isEdge(neighbors.SE), isEdge(neighbors.S))
	drawSection(d, e, isEdge(neighbors.S), isEdge(neighbors.NE), isEdge(neighbors.SE))
	drawSection(e, f, isEdge(neighbors.SE), isEdge(neighbors.N), isEdge(neighbors.NE))
	drawSection(f, a, isEdge(neighbors.NE), isEdge(neighbors.NW), isEdge(neighbors.N))
}

func setEdgeNormal(edge bool) {
	if edge {
		rl.Normal3f(1.0, 0.0, 0.0)
		return
	}
	rl.Normal3f(0.0, 0.0, 0.0)
}

func drawHexagonBuffered(x, y float32, size vec.Vec2, color color.RGBA) {
	width := size.X * 2.0
	height := size.Y * sqrt3
	wp := width / 4.0
	hp := height / 2.0
	ox := width / 2.0
	oy := hp
	a := rl.Vector2{X: x - ox + wp, Y: y - oy}
	b := rl.Vector2{X: x - ox, Y: y - oy + hp}
	c := rl.Vector2{X: x - ox + wp, Y: y - oy + height}
	d := rl.Vector2{X: x - ox + wp*3, Y: y - oy + height}
	e := rl.Vector2{X: x - ox + width, Y: y - oy + hp}
	f := rl.Vector2{X: x - ox + wp*3, Y: y - oy}
	center := rl.Vector2{X: x, Y: y}

	rl.Color4ub(color.R, color.G, color.B, color.A)
	rl.Vertex2f(a.X, a.Y)
	rl.Vertex2f(b.X, b.Y)
	rl.Vertex2f(center.X, center.Y)
	rl.Vertex2f(b.X, b.Y)
	rl.Vertex2f(c.X, c.Y)
	rl.Vertex2f(center.X, center.Y)
	rl.Vertex2f(c.X, c.Y)
	rl.Vertex2f(d.X, d.Y)
	rl.Vertex2f(center.X, center.Y)
	rl.Vertex2f(d.X, d.Y)
	rl.Vertex2f(e.X, e.Y)
	rl.Vertex2f(center.X, center.Y)
	rl.Vertex2f(e.X, e.Y)
	rl.Vertex2f(f.X, f.Y)
	rl.Vertex2f(center.X, center.Y)
	rl.Vertex2f(f.X, f.Y)
	rl.Vertex2f(a.X, a.Y)
	rl.Vertex2f(center.X, center.Y)
}
