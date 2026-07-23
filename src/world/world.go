package world

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/global"
	"github.com/threeidiotsonegamejam/gmtk26/src/mathutil"
	v "github.com/threeidiotsonegamejam/gmtk26/src/mathutil/vec"
)

type World struct {
	Generator     Generator
	Grid          [][]Cell
	TileToGrid    map[string][]v.Vec2i
	GridSize      v.Vec2i
	Camera        rl.Camera2D
	HexSize       v.Vec2
	HasInit       bool
	BGShader      rl.Shader
	BGTimeLoc     int32
	VoidShader    rl.Shader
	PanStart      v.Vec2
	Viewport      rl.RenderTexture2D
	MousePosition v.Vec2
}

var sqrt3 = float32(math.Sqrt(3.0))

func (w *World) Init() {
	w.HasInit = true

	if w.Camera.Zoom == 0.0 {
		w.Camera.Zoom = 1.0
	}
	if w.HexSize == (v.Vec2{}) {
		w.HexSize = v.Vec2{X: 48.0, Y: 48.0}
	}
	if w.GridSize == (v.Vec2i{}) {
		w.GridSize = v.Vec2i{X: 100, Y: 70}
	}

	w.Grid = w.Generator.Generate(w.GridSize)

	w.TileToGrid = make(map[string][]v.Vec2i)

	for x := range w.GridSize.X {
		for y := range w.GridSize.Y {
			cell := w.Grid[x][y]
			tileData := cell.Tile.Data()
			if w.TileToGrid[tileData.Type] == nil {
				w.TileToGrid[tileData.Type] = make([]v.Vec2i, 0)
			}

			w.TileToGrid[tileData.Type] = append(w.TileToGrid[tileData.Type], v.Vec2i{X: int32(x), Y: int32(y)})
		}
	}

	// FIXME: DEATH THIS IS DEATH!!! WELL, AT LEAST UNTIL WE ADD SOMETHING TO UNLOAD IT...
	w.BGShader = rl.LoadShader("assets/shaders/base.vert", "assets/shaders/bg.frag")
	w.BGTimeLoc = rl.GetLocationUniform(w.BGShader.ID, "time")
	w.VoidShader = rl.LoadShader("assets/shaders/base.vert", "assets/shaders/void.frag")

	w.Viewport = rl.LoadRenderTexture(global.ViewportSize.X, global.ViewportSize.Y)
}

func (w *World) Update(delta float32) {
	if w.Viewport.Texture.Width != global.ViewportSize.X {
		rl.UnloadRenderTexture(w.Viewport)
		w.Viewport = rl.LoadRenderTexture(global.ViewportSize.X, global.ViewportSize.Y)
	}

	mousePos := v.Vec2FromRL(rl.GetScreenToWorld2D(rl.Vector2(w.MousePosition), w.Camera))
	if rl.IsMouseButtonPressed(rl.MouseButtonRight) {
		w.PanStart = mousePos
	}

	if rl.IsMouseButtonDown(rl.MouseButtonRight) {
		mouseDelta := w.PanStart.Sub(mousePos)
		w.Camera.Target = w.Camera.Target.Add(mouseDelta.ToRL())
	}

	moveDir := v.Vec2{}
	if rl.IsKeyDown(rl.KeyW) {
		moveDir.Y -= 1.0
	}
	if rl.IsKeyDown(rl.KeyA) {
		moveDir.X -= 1.0
	}

	if rl.IsKeyDown(rl.KeyS) {
		moveDir.Y += 1.0
	}

	if rl.IsKeyDown(rl.KeyD) {
		moveDir.X += 1.0
	}

	w.Camera.Target = v.Vec2FromRL(w.Camera.Target).Add(moveDir.Normalize().Mul(v.Vec2{X: 10.0, Y: 10.0})).ToRL()

	w.Camera.Offset.X = float32(global.ViewportSize.X) / 2.0
	w.Camera.Offset.Y = float32(global.ViewportSize.Y) / 2.0

	w.Camera.Zoom += rl.GetMouseWheelMove() * 0.5

	if w.Camera.Zoom > 5.0 {
		w.Camera.Zoom = 5.0
	} else if w.Camera.Zoom < 0.08 {
		w.Camera.Zoom = 0.08
	}

	w.Camera.Target = v.Vec2FromRL(w.Camera.Target).Round().ToRL()
}

func (w *World) Draw() {
	screenW := float32(rl.GetRenderWidth())
	screenH := float32(rl.GetRenderHeight())

	viewW := float32(w.Viewport.Texture.Width)
	viewH := float32(w.Viewport.Texture.Height)

	srcRect := rl.Rectangle{
		X:      0.0,
		Y:      0.0,
		Width:  viewW,
		Height: -viewH,
	}
	ratio := float32(int32(math.Min(
		float64(screenW/viewW),
		float64(screenH/viewH),
	))) + 1
	dstRect := rl.Rectangle{
		X:      (screenW - viewW*ratio) / 2.0,
		Y:      (screenH - viewH*ratio) / 2.0,
		Width:  viewW * ratio,
		Height: viewH * ratio,
	}
	mouse := global.MousePosition
	w.MousePosition = v.Vec2{
		X: (mouse.X - dstRect.X) * (srcRect.Width / dstRect.Width),
		Y: (mouse.Y - dstRect.Y) * (-srcRect.Height / dstRect.Height),
	}

	rl.BeginTextureMode(w.Viewport)

	mousePos := v.Vec2FromRL(rl.GetScreenToWorld2D(rl.Vector2(w.MousePosition), w.Camera))
	if rl.IsShaderValid(w.BGShader) {
		rl.SetShaderValue(w.BGShader, w.BGTimeLoc, []float32{float32(rl.GetTime())}, rl.ShaderUniformFloat)
		rl.BeginShaderMode(w.BGShader)
		rl.Begin(rl.Triangles)

		width, height := float32(rl.GetRenderWidth()), float32(rl.GetRenderHeight())

		rl.Color4ub(255, 255, 0, 255)
		rl.Normal3f(0.0, 0.0, 1.0)

		rl.TexCoord2f(0.0, 0.0)
		rl.Vertex2f(0, 0)

		rl.TexCoord2f(width, height)
		rl.Vertex2f(width, height)

		rl.TexCoord2f(width, 0.0)
		rl.Vertex2f(width, 0)

		rl.TexCoord2f(0.0, height)
		rl.Vertex2f(0, height)

		rl.TexCoord2f(width, height)
		rl.Vertex2f(width, height)

		rl.TexCoord2f(0.0, 0.0)
		rl.Vertex2f(0, 0)

		rl.End()
		rl.EndShaderMode()
	}
	if rl.IsShaderValid(w.VoidShader) {
		timeLoc := rl.GetLocationUniform(w.VoidShader.ID, "time")
		rl.SetShaderValue(w.VoidShader, timeLoc, []float32{float32(rl.GetTime())}, rl.ShaderUniformFloat)
	}

	rl.BeginMode2D(w.Camera)

	topLeft := rl.GetScreenToWorld2D(rl.Vector2{}, w.Camera)
	topLeft = topLeft.Subtract(rl.Vector2{X: w.HexSize.X * 2.0, Y: w.HexSize.Y * 2.0})
	bottomRight := rl.GetScreenToWorld2D(rl.Vector2{X: float32(rl.GetRenderWidth()), Y: float32(rl.GetRenderHeight())}, w.Camera)
	bottomRight = bottomRight.Add(rl.Vector2{X: w.HexSize.X * 2.0, Y: w.HexSize.Y * 2.0})

	width := w.HexSize.X * 2.0
	height := w.HexSize.Y * sqrt3

	rl.Begin(rl.Triangles)
	for x := range len(w.Grid) {
		for y, cell := range w.Grid[x] {
			yOffset := float32(height/2.0) * float32(x%2)
			worldPos := v.Vec2{X: float32(x) * width / 4.0 * 3.0, Y: float32(y)*height + yOffset}
			if worldPos.X < topLeft.X || worldPos.X > bottomRight.X || worldPos.Y < topLeft.Y || worldPos.Y > bottomRight.Y {
				continue
			}
			tileData := cell.Tile.Data()

			hex := w.PixelToHex(mousePos)

			tileColor := tileData.Color
			if hex.X == int32(x) && hex.Y == int32(y) {
				tileColor = *mathutil.ColorAdd(tileColor, 20)
			}

			DrawHexagonBuffered(worldPos.X, worldPos.Y, w.HexSize, tileColor)
		}
	}
	rl.End()
	for _, tiles := range w.TileToGrid {
		for i, tilePos := range tiles {
			yOffset := float32(height/2.0) * float32(tilePos.X%2)
			worldPos := v.Vec2{X: float32(tilePos.X) * width / 4.0 * 3.0, Y: float32(tilePos.Y)*height + yOffset}
			if worldPos.X < topLeft.X || worldPos.X > bottomRight.X || worldPos.Y < topLeft.Y || worldPos.Y > bottomRight.Y {
				if i == 0 {
					w.GetCell(tilePos).Tile.Draw(w, worldPos, tilePos, DrawStateBegin)
				} else if i == len(tiles)-1 {
					w.GetCell(tilePos).Tile.Draw(w, worldPos, tilePos, DrawStateEnd)
				}
				continue
			}
			tile := w.GetCell(tilePos).Tile

			drawState := DrawStateNormal
			if i == 0 {
				drawState = DrawStateBegin
			} else if i == len(tiles)-1 {
				drawState = DrawStateEnd
			}
			tile.Draw(w, worldPos, tilePos, drawState)
		}
	}

	rl.EndMode2D()
	rl.EndTextureMode()

	rl.DrawTexturePro(w.Viewport.Texture, srcRect, dstRect, rl.Vector2{}, 0.0, rl.White)
}

type Neighbors struct {
	NW *Cell
	N  *Cell
	NE *Cell
	SE *Cell
	S  *Cell
	SW *Cell
}

func (w World) GetNeighbors(pos v.Vec2i) Neighbors {
	if pos.X%2 == 0 {
		return Neighbors{
			NW: w.GetCell(pos.Add(v.Vec2i{X: -1, Y: -1})),
			N:  w.GetCell(pos.Add(v.Vec2i{X: 0, Y: -1})),
			NE: w.GetCell(pos.Add(v.Vec2i{X: 1, Y: -1})),
			SW: w.GetCell(pos.Add(v.Vec2i{X: -1, Y: 0})),
			S:  w.GetCell(pos.Add(v.Vec2i{X: 0, Y: 1})),
			SE: w.GetCell(pos.Add(v.Vec2i{X: 1, Y: 0})),
		}
	}
	return Neighbors{
		NW: w.GetCell(pos.Add(v.Vec2i{X: -1, Y: 0})),
		N:  w.GetCell(pos.Add(v.Vec2i{X: 0, Y: -1})),
		NE: w.GetCell(pos.Add(v.Vec2i{X: 1, Y: 0})),
		SW: w.GetCell(pos.Add(v.Vec2i{X: -1, Y: 1})),
		S:  w.GetCell(pos.Add(v.Vec2i{X: 0, Y: 1})),
		SE: w.GetCell(pos.Add(v.Vec2i{X: 1, Y: 1})),
	}
}

func (w World) GetCell(pos v.Vec2i) *Cell {
	if pos.X < 0 || pos.X >= w.GridSize.X || pos.Y < 0 || pos.Y >= w.GridSize.Y {
		return nil
	}
	return &w.Grid[pos.X][pos.Y]
}

func DrawHexagon(x float32, y float32, size v.Vec2, color rl.Color) {
	rl.Begin(rl.Triangles)
	DrawHexagonBuffered(x, y, size, color)
	rl.End()
}

func DrawHexagonBuffered(x float32, y float32, size v.Vec2, color rl.Color) {
	w := size.X * 2.0
	h := size.Y * sqrt3
	wp := w / 4.0
	hp := h / 2.0
	ox := w / 2.0
	oy := hp

	a := rl.Vector2{X: x - ox + wp, Y: y - oy}
	b := rl.Vector2{X: x - ox, Y: y - oy + hp}
	c := rl.Vector2{X: x - ox + wp, Y: y - oy + h}
	d := rl.Vector2{X: x - ox + wp*3, Y: y - oy + h}
	e := rl.Vector2{X: x - ox + w, Y: y - oy + hp}
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
