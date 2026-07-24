package world

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/global"
	"github.com/threeidiotsonegamejam/gmtk26/src/util"
	v "github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

type World struct {
	Seed          int64
	Grid          [][]Cell
	TileToGrid    map[string][]v.Vec2i
	GridSize      v.Vec2i
	Camera        rl.Camera2D
	ZoomAnchor    v.Vec2
	TargetZoom    float32
	HexSize       v.Vec2
	HasInit       bool
	BGShader      rl.Shader
	BGTimeLoc     int32
	VoidShader    rl.Shader
	PanStart      v.Vec2
	PanVelocity   v.Vec2
	Viewport      rl.RenderTexture2D
	MousePosition v.Vec2
}

var sqrt3 = float32(math.Sqrt(3.0))

const (
	// Number of nanoseconds in one second. Update receives delta in nanoseconds,
	// while movement and smoothing calculations expect seconds.
	nanosecondsPerSecond float32 = 1_000_000_000.0

	// How far the camera may show beyond the world edge, measured in screen
	// pixels. Set to 0.0 to prevent showing anything outside the world.
	cameraBoundsPadding float32 = 200.0

	// Stops momentum on an axis when the camera reaches that axis's boundary.
	// Disable this if you prefer momentum to continue pushing against the edge.
	stopMomentumAtBounds = false

	// Base WASD camera movement speed in world units per second.
	cameraMoveSpeed float32 = 1000.0

	// Controls how strongly WASD movement changes with zoom.
	// 0.0 = movement ignores zoom.
	// 0.5 = movement is divided by sqrt(zoom).
	// 1.0 = movement is divided directly by zoom.
	cameraMoveZoomExponent float64 = 0.5

	// Smallest and largest allowed camera zoom values.
	cameraMinZoom float32 = 0.08
	cameraMaxZoom float32 = 1.0

	// Amount each mouse-wheel step changes the target zoom.
	// Higher values make each wheel step zoom farther.
	cameraZoomStep float32 = 0.18

	// How quickly the actual zoom approaches the target zoom.
	// Lower values feel softer; higher values feel more immediate.
	cameraZoomSmoothness float32 = 12.0

	// The momentum speed that slow and fast mouse drags are pulled toward.
	// This represents approximate screen-space movement per second.
	panMomentumMidSpeed float32 = 600.0

	// Controls how much the original drag speed affects momentum.
	// 0.0 = every drag produces the midpoint speed.
	// 0.35 = strong compression while preserving some speed variation.
	// 1.0 = momentum directly matches the original drag speed.
	panSpeedCompressionExponent float64 = 0.4

	// How much new drag velocity is mixed into the existing momentum.
	// Higher values react faster to the latest mouse movement.
	panVelocitySampleWeight float32 = 0.35

	// How quickly momentum slows after releasing the mouse button.
	// Lower values glide longer; higher values stop sooner.
	panMomentumDamping float32 = 4.0
)

func (w *World) Init() {
	w.HasInit = true

	if w.Camera.Zoom == 0.0 {
		w.Camera.Zoom = 0.75
	}
	if w.TargetZoom == 0.0 {
		w.TargetZoom = w.Camera.Zoom
	}
	if w.HexSize == (v.Vec2{}) {
		w.HexSize = v.Vec2{X: 48.0, Y: 48.0}
	}
	if w.GridSize == (v.Vec2i{}) {
		w.GridSize = v.Vec2i{X: 96, Y: 96}
	}
	w.Camera.Target = w.GridSize.Vec2().Mul(w.HexSize).Sub(global.ViewportSize.Vec2()).ToRL()

	w.Generate()

	// FIXME: DEATH THIS IS DEATH!!! WELL, AT LEAST UNTIL WE ADD SOMETHING TO UNLOAD IT...
	w.BGShader = rl.LoadShader("assets/shaders/base.vert", "assets/shaders/bg.frag")
	w.BGTimeLoc = rl.GetLocationUniform(w.BGShader.ID, "time")
	w.VoidShader = rl.LoadShader("assets/shaders/base.vert", "assets/shaders/void.frag")

	w.Viewport = rl.LoadRenderTexture(global.ViewportSize.X, global.ViewportSize.Y)
}

func (w *World) Generate() {
	w.Grid = Generate(w.GridSize, w.Seed)
	w.TileToGrid = make(map[string][]v.Vec2i)

	for x := range w.GridSize.X {
		for y := range w.GridSize.Y {
			cell := w.Grid[x][y]
			tile := cell.Tile
			if w.TileToGrid[tile.Type] == nil {
				w.TileToGrid[tile.Type] = make([]v.Vec2i, 0)
			}

			w.TileToGrid[tile.Type] = append(w.TileToGrid[tile.Type], v.Vec2i{X: int32(x), Y: int32(y)})
		}
	}

	SpreadResources(w, w.Seed)
}

func (w *World) ClampCameraToWorld() {
	hexWidth := w.HexSize.X * 2.0
	hexHeight := w.HexSize.Y * sqrt3

	// These bounds include the full size of the outermost hexagons.
	worldMinX := -w.HexSize.X
	worldMaxX := float32(w.GridSize.X-1)*hexWidth/4.0*3.0 + w.HexSize.X
	worldMinY := -hexHeight / 2.0
	worldMaxY := float32(w.GridSize.Y) * hexHeight

	// Convert the screen-space padding into world-space padding so it looks
	// approximately the same at every zoom level.
	worldPadding := cameraBoundsPadding / w.Camera.Zoom

	worldMinX -= worldPadding
	worldMaxX += worldPadding
	worldMinY -= worldPadding
	worldMaxY += worldPadding

	// Camera.Target is positioned at Camera.Offset. Since the offset is the
	// center of the viewport, these are the visible world-space half sizes.
	viewHalfWidth := float32(w.Viewport.Texture.Width) / (2.0 * w.Camera.Zoom)
	viewHalfHeight := float32(w.Viewport.Texture.Height) / (2.0 * w.Camera.Zoom)

	minTargetX := worldMinX + viewHalfWidth
	maxTargetX := worldMaxX - viewHalfWidth
	minTargetY := worldMinY + viewHalfHeight
	maxTargetY := worldMaxY - viewHalfHeight

	target := v.Vec2FromRL(w.Camera.Target)
	previousTarget := target

	// When zoomed out far enough that the viewport is larger than the world,
	// there is no valid clamp range. Center the camera on that axis instead.
	if minTargetX > maxTargetX {
		target.X = (worldMinX + worldMaxX) / 2.0
	} else {
		target.X = max(minTargetX, min(target.X, maxTargetX))
	}

	if minTargetY > maxTargetY {
		target.Y = (worldMinY + worldMaxY) / 2.0
	} else {
		target.Y = max(minTargetY, min(target.Y, maxTargetY))
	}

	w.Camera.Target = target.ToRL()

	if stopMomentumAtBounds {
		if target.X != previousTarget.X {
			w.PanVelocity.X = 0.0
		}
		if target.Y != previousTarget.Y {
			w.PanVelocity.Y = 0.0
		}
	}
}

func (w *World) Update(delta float32) {
	deltaSeconds := delta / nanosecondsPerSecond

	screenW := float32(rl.GetRenderWidth())
	screenH := float32(rl.GetRenderHeight())
	viewH := float32(w.Viewport.Texture.Height)
	ratio := screenH / viewH
	viewW := float32(int32(screenW/ratio) + 1)
	srcRect := rl.Rectangle{X: 0, Y: 0, Width: viewW, Height: -viewH}
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

	w.Camera.Offset.X = float32(w.Viewport.Texture.Width) / 2.0
	w.Camera.Offset.Y = float32(w.Viewport.Texture.Height) / 2.0

	mousePos := v.Vec2FromRL(rl.GetScreenToWorld2D(rl.Vector2(w.MousePosition), w.Camera))

	hex := w.PixelToHex(mousePos)
	if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
		fmt.Printf("Clicked cell: x=%d, y=%d\n", hex.X, hex.Y)
	}

	if rl.IsMouseButtonPressed(rl.MouseButtonRight) {
		w.PanStart = w.MousePosition
		w.PanVelocity = v.Vec2{}
	}

	if rl.IsMouseButtonDown(rl.MouseButtonRight) {
		mouseDelta := w.MousePosition.Sub(w.PanStart)
		panDelta := mouseDelta.Mul(v.Vec2{
			X: -1.0 / w.Camera.Zoom,
			Y: -1.0 / w.Camera.Zoom,
		})

		w.Camera.Target = v.Vec2FromRL(w.Camera.Target).Add(panDelta).ToRL()

		if deltaSeconds > 0.0 {
			rawVelocity := panDelta.Mul(v.Vec2{
				X: 1.0 / deltaSeconds,
				Y: 1.0 / deltaSeconds,
			})

			rawSpeed := float32(math.Sqrt(float64(
				rawVelocity.X*rawVelocity.X +
					rawVelocity.Y*rawVelocity.Y,
			)))

			if rawSpeed > 0.0 {
				// Convert the desired screen-space momentum speed to world-space
				// speed so momentum feels similar at different zoom levels.
				midSpeed := panMomentumMidSpeed / w.Camera.Zoom

				// Compress very slow and very fast drags toward midSpeed while
				// retaining some variation from the original drag speed.
				compressedSpeed := midSpeed * float32(math.Pow(
					float64(rawSpeed/midSpeed),
					panSpeedCompressionExponent,
				))

				dragVelocity := rawVelocity.Mul(v.Vec2{
					X: compressedSpeed / rawSpeed,
					Y: compressedSpeed / rawSpeed,
				})

				previousVelocityWeight := 1.0 - panVelocitySampleWeight
				w.PanVelocity = w.PanVelocity.Mul(v.Vec2{
					X: previousVelocityWeight,
					Y: previousVelocityWeight,
				}).Add(dragVelocity.Mul(v.Vec2{
					X: panVelocitySampleWeight,
					Y: panVelocitySampleWeight,
				}))
			}
		}

		w.PanStart = w.MousePosition
	} else {
		w.Camera.Target = v.Vec2FromRL(w.Camera.Target).Add(w.PanVelocity.Mul(v.Vec2{
			X: deltaSeconds,
			Y: deltaSeconds,
		})).ToRL()

		// Exponential damping makes momentum frame-rate independent.
		panDecay := float32(math.Exp(float64(-panMomentumDamping * deltaSeconds)))
		w.PanVelocity = w.PanVelocity.Mul(v.Vec2{
			X: panDecay,
			Y: panDecay,
		})
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

	if moveDir.X != 0.0 || moveDir.Y != 0.0 {
		zoomMovementScale := float32(math.Pow(
			float64(w.Camera.Zoom),
			cameraMoveZoomExponent,
		))
		moveDistance := cameraMoveSpeed * deltaSeconds / zoomMovementScale
		w.Camera.Target = v.Vec2FromRL(w.Camera.Target).Add(moveDir.Normalize().Mul(v.Vec2{
			X: moveDistance,
			Y: moveDistance,
		})).ToRL()
	}

	wheel := rl.GetMouseWheelMove()
	if wheel != 0.0 {
		// Store the cursor position so the same world point remains underneath
		// it while the camera smoothly approaches the new zoom.
		w.ZoomAnchor = w.MousePosition
		w.TargetZoom *= float32(math.Exp(float64(wheel * cameraZoomStep)))
	}

	if w.TargetZoom > cameraMaxZoom {
		w.TargetZoom = cameraMaxZoom
	} else if w.TargetZoom < cameraMinZoom {
		w.TargetZoom = cameraMinZoom
	}

	zoomBlend := 1.0 - float32(math.Exp(float64(-cameraZoomSmoothness*deltaSeconds)))
	nextZoom := w.Camera.Zoom + (w.TargetZoom-w.Camera.Zoom)*zoomBlend

	if nextZoom != w.Camera.Zoom {
		// Find the world position under the cursor before zooming.
		zoomBefore := v.Vec2FromRL(rl.GetScreenToWorld2D(rl.Vector2(w.ZoomAnchor), w.Camera))

		w.Camera.Zoom = nextZoom

		// Move the camera target by the difference after zooming, keeping the
		// same world position locked beneath the cursor.
		zoomAfter := v.Vec2FromRL(rl.GetScreenToWorld2D(rl.Vector2(w.ZoomAnchor), w.Camera))
		w.Camera.Target = v.Vec2FromRL(w.Camera.Target).Add(zoomBefore.Sub(zoomAfter)).ToRL()
	}

	w.ClampCameraToWorld()
}

func (w *World) Draw() {
	screenW := float32(rl.GetRenderWidth())
	screenH := float32(rl.GetRenderHeight())
	viewH := float32(w.Viewport.Texture.Height)
	ratio := screenH / viewH
	viewW := float32(int32(screenW/ratio) + 1)

	if w.Viewport.Texture.Width != int32(viewW) {
		rl.UnloadRenderTexture(w.Viewport)
		w.Viewport = rl.LoadRenderTexture(int32(viewW), int32(viewH))
	}

	srcRect := rl.Rectangle{
		X:      0.0,
		Y:      0.0,
		Width:  viewW,
		Height: -viewH,
	}
	dstRect := rl.Rectangle{
		X:      (screenW - viewW*ratio) / 2.0,
		Y:      (screenH - viewH*ratio) / 2.0,
		Width:  viewW * ratio,
		Height: viewH * ratio,
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
			tile := cell.Tile

			hex := w.PixelToHex(mousePos)

			tileColor := tile.Color
			if hex.X == int32(x) && hex.Y == int32(y) {
				tileColor = *util.ColorAdd(tileColor, 30)
			}
			if x%2 == 1 {
				tileColor = *util.ColorSub(tileColor, 12)
			}
			if y%2 == 1 {
				tileColor = *util.ColorSub(tileColor, 6)
			}

			DrawHexagonBuffered(worldPos.X, worldPos.Y, w.HexSize, tileColor)
		}
	}
	rl.End()
	for _, tiles := range w.TileToGrid {
		for i, tilePos := range tiles {
			tile := w.GetCell(tilePos).Tile
			if tile.Draw == nil {
				continue
			}
			yOffset := float32(height/2.0) * float32(tilePos.X%2)
			worldPos := v.Vec2{X: float32(tilePos.X) * width / 4.0 * 3.0, Y: float32(tilePos.Y)*height + yOffset}
			if worldPos.X < topLeft.X || worldPos.X > bottomRight.X || worldPos.Y < topLeft.Y || worldPos.Y > bottomRight.Y {
				if i == 0 {
					tile.Draw(*w, tilePos, worldPos, DrawStateBegin)
				} else if i == len(tiles)-1 {
					tile.Draw(*w, tilePos, worldPos, DrawStateEnd)
				}
				continue
			}

			drawState := DrawStateNormal
			if i == 0 {
				drawState = DrawStateBegin
			} else if i == len(tiles)-1 {
				drawState = DrawStateEnd
			}
			tile.Draw(*w, tilePos, worldPos, drawState)
		}
	}
	for x := range len(w.Grid) {
		for y, cell := range w.Grid[x] {
			yOffset := float32(height/2.0) * float32(x%2)
			worldPos := v.Vec2{X: float32(x) * width / 4.0 * 3.0, Y: float32(y)*height + yOffset}
			if worldPos.X < topLeft.X || worldPos.X > bottomRight.X || worldPos.Y < topLeft.Y || worldPos.Y > bottomRight.Y {
				continue
			}

			color := rl.Black

			switch cell.Resource {
			case game.ResourceUnknown:
				continue
			case game.ResourceGold:
				color = rl.Gold
			case game.ResourceCoal:
				color = rl.Black
			case game.ResourceIron:
				color = rl.ColorLerp(rl.Brown, rl.White, 0.6)
			case game.ResourceWood:
				color = rl.Brown
			}

			rl.DrawRectangle(int32(worldPos.X)-16, int32(worldPos.Y)-16, 32, 32, color)
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
