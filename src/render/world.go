package render

import (
	"math"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/global"
	"github.com/threeidiotsonegamejam/gmtk26/src/render/shaders"
	"github.com/threeidiotsonegamejam/gmtk26/src/util"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/rlutil"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/rlvec"
	v "github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

var sqrt3 = float32(math.Sqrt(3.0))

const (
	// Extra screen-space distance the camera may show beyond the map.
	cameraBoundsPadding float32 = 200.0

	// Stop momentum on an axis when it reaches the map boundary.
	stopMomentumAtBounds = false

	// Base WASD movement speed in world units per second.
	cameraMoveSpeed float32 = 1000.0

	// How strongly zoom reduces WASD speed: 0 ignores zoom, 1 scales directly.
	cameraMoveZoomExponent = 0.5

	cameraMinZoom     float32 = 0.08
	cameraMaxZoom     float32 = 1.0
	cameraDefaultZoom float32 = 0.75

	// Zoom change per wheel step and speed of interpolation toward it.
	cameraZoomStep       float32 = 0.18
	cameraZoomSmoothness float32 = 12.0

	// Approximate screen-space momentum speed that drags converge toward.
	panMomentumMidSpeed float32 = 600.0

	// Compresses extreme drag speeds while retaining some variation.
	panSpeedCompressionExponent = 0.4

	// How much each new drag sample replaces existing momentum; higher values
	// react faster, while lower values produce smoother momentum.
	panVelocitySampleWeight = float32(0.35)

	// Rate at which momentum decays after releasing the mouse.
	panMomentumDamping = float32(4.0)

	// Smooth camera movement toward focus target.
	cameraFocusSmoothness float32 = 8.0
)

type WorldRenderer struct {
	Camera        rl.Camera2D
	ZoomAnchor    v.Vec2
	TargetZoom    float32
	HexSize       v.Vec2
	PanStart      v.Vec2
	PanVelocity   v.Vec2
	MousePosition v.Vec2

	HoveredHex  game.Hex
	SelectedHex *game.Hex

	TargetPosition   v.Vec2
	InterpolateFocus bool

	BuildingToPlace game.BuildingType
	// OnPlaceBuilding, when set, may take over a placement click. Returning
	// true means the click was handled externally (e.g. sent to the server)
	// and the map is not modified locally.
	OnPlaceBuilding func(hex game.Hex, building game.BuildingType) bool
	buildingPreview buildingPreview

	viewport    rl.RenderTexture2D
	initialized bool
}

func (r *WorldRenderer) Init(m *game.Map) {
	if r.initialized {
		return
	}
	r.initialized = true

	if r.HexSize == (v.Vec2{}) {
		r.HexSize = v.Vec2{X: 48.0, Y: 48.0}
	}
	r.ResetCamera(m)

	r.viewport = rl.LoadRenderTexture(global.ViewportSize.X, global.ViewportSize.Y)
}

func (r *WorldRenderer) ResetCamera(m *game.Map) {
	if r.HexSize == (v.Vec2{}) {
		r.HexSize = v.Vec2{X: 48.0, Y: 48.0}
	}

	r.Camera.Zoom = cameraDefaultZoom
	r.TargetZoom = cameraDefaultZoom
	r.PanVelocity = v.Vec2{}
	r.Camera.Target = rlvec.ToRL(v.Vec2{
		X: float32(m.GridSize.X),
		Y: float32(m.GridSize.Y),
	}.Mul(r.HexSize).Sub(global.ViewportSize.Vec2()))
	r.Camera.Offset = rlvec.ToRL(global.ViewportSize.Vec2().Mul(v.Vec2{X: 0.5, Y: 0.5}))
	r.TargetPosition = rlvec.FromRL(r.Camera.Target)
}

func (r *WorldRenderer) Unload() {
	if !r.initialized {
		return
	}
	rl.UnloadRenderTexture(r.viewport)
	r.initialized = false
}

func (r *WorldRenderer) Update(m *game.Map, delta time.Duration) {
	deltaSeconds := float32(delta.Seconds())
	srcRect, dstRect := r.viewportRects()
	mouse := global.MousePosition
	// Convert window coordinates into the fixed-height render viewport.
	r.MousePosition = v.Vec2{
		X: (mouse.X - dstRect.X) * (srcRect.Width / dstRect.Width),
		Y: (mouse.Y - dstRect.Y) * (-srcRect.Height / dstRect.Height),
	}

	r.Camera.Offset.X = float32(r.viewport.Texture.Width) / 2.0
	r.Camera.Offset.Y = float32(r.viewport.Texture.Height) / 2.0

	if global.UIBlocksWorldInput {
		r.clampCameraToMap(m)
		mousePos := rlvec.FromRL(rl.GetScreenToWorld2D(rl.Vector2(r.MousePosition), r.Camera))
		r.updateBuildingPlacement(m, r.PixelToHex(mousePos))
	} else {
		if rl.IsMouseButtonPressed(rl.MouseButtonRight) {
			r.PanStart = r.MousePosition
			r.PanVelocity = v.Vec2{}
			r.InterpolateFocus = false
		}

		if rl.IsMouseButtonDown(rl.MouseButtonRight) {
			mouseDelta := r.MousePosition.Sub(r.PanStart)
			panDelta := mouseDelta.Mul(v.Vec2{
				X: -1.0 / r.Camera.Zoom,
				Y: -1.0 / r.Camera.Zoom,
			})

			r.Camera.Target = rlvec.ToRL(rlvec.FromRL(r.Camera.Target).Add(panDelta))

			if deltaSeconds > 0.0 {
				rawVelocity := panDelta.Mul(v.Vec2{
					X: 1.0 / deltaSeconds,
					Y: 1.0 / deltaSeconds,
				})
				rawSpeed := rawVelocity.Magnitude()

				if rawSpeed > 0.0 {
					midSpeed := panMomentumMidSpeed / r.Camera.Zoom
					compressedSpeed := midSpeed * float32(math.Pow(
						float64(rawSpeed/midSpeed),
						panSpeedCompressionExponent,
					))
					dragVelocity := rawVelocity.Mul(v.Vec2{
						X: compressedSpeed / rawSpeed,
						Y: compressedSpeed / rawSpeed,
					})
					previousVelocityWeight := 1.0 - panVelocitySampleWeight
					r.PanVelocity = r.PanVelocity.Mul(v.Vec2{
						X: previousVelocityWeight,
						Y: previousVelocityWeight,
					}).Add(dragVelocity.Mul(v.Vec2{
						X: panVelocitySampleWeight,
						Y: panVelocitySampleWeight,
					}))
				}
			}

			r.PanStart = r.MousePosition
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
			zoomMovementScale := float32(math.Pow(float64(r.Camera.Zoom), cameraMoveZoomExponent))
			moveDistance := cameraMoveSpeed * deltaSeconds / zoomMovementScale
			r.Camera.Target = rlvec.ToRL(rlvec.FromRL(r.Camera.Target).Add(moveDir.Normalize().Mul(v.Vec2{
				X: moveDistance,
				Y: moveDistance,
			})))
			r.InterpolateFocus = false
		}

		wheel := rl.GetMouseWheelMove()
		if wheel != 0.0 {
			r.ZoomAnchor = r.MousePosition
			r.TargetZoom *= float32(math.Exp(float64(wheel * cameraZoomStep)))
		}
	}

	if !rl.IsMouseButtonDown(rl.MouseButtonRight) {
		r.Camera.Target = rlvec.ToRL(rlvec.FromRL(r.Camera.Target).Add(r.PanVelocity.Mul(v.Vec2{
			X: deltaSeconds,
			Y: deltaSeconds,
		})))

		panDecay := float32(math.Exp(float64(-panMomentumDamping * deltaSeconds)))
		r.PanVelocity = r.PanVelocity.Mul(v.Vec2{X: panDecay, Y: panDecay})
	}

	r.TargetZoom = max(cameraMinZoom, min(r.TargetZoom, cameraMaxZoom))
	zoomBlend := 1.0 - float32(math.Exp(float64(-cameraZoomSmoothness*deltaSeconds)))
	nextZoom := r.Camera.Zoom + (r.TargetZoom-r.Camera.Zoom)*zoomBlend

	if nextZoom != r.Camera.Zoom {
		// Offset the camera by the cursor's world-space shift after zooming.
		zoomBefore := rlvec.FromRL(rl.GetScreenToWorld2D(rl.Vector2(r.ZoomAnchor), r.Camera))
		r.Camera.Zoom = nextZoom
		zoomAfter := rlvec.FromRL(rl.GetScreenToWorld2D(rl.Vector2(r.ZoomAnchor), r.Camera))
		r.Camera.Target = rlvec.ToRL(rlvec.FromRL(r.Camera.Target).Add(zoomBefore.Sub(zoomAfter)))
	}

	if r.InterpolateFocus {
		currentTarget := rlvec.FromRL(r.Camera.Target)
		diff := r.TargetPosition.Sub(currentTarget)
		if diff.MagnitudeSqr() > 0.01 {
			focusBlend := 1.0 - float32(math.Exp(float64(-cameraFocusSmoothness*deltaSeconds)))
			r.Camera.Target = rlvec.ToRL(currentTarget.Lerp(r.TargetPosition, focusBlend))
		} else {
			r.Camera.Target = rlvec.ToRL(r.TargetPosition)
			r.InterpolateFocus = false
		}
	}

	r.clampCameraToMap(m)

	mousePos := rlvec.FromRL(rl.GetScreenToWorld2D(rl.Vector2(r.MousePosition), r.Camera))
	hex := r.PixelToHex(mousePos)
	r.HoveredHex = hex
	r.updateBuildingPlacement(m, hex)
	if !global.UIBlocksWorldInput && r.BuildingToPlace == game.BuildingUnknown {
		if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
			cell := m.GetCell(hex)
			if cell != nil && cell.Building != game.BuildingUnknown {
				h := hex
				r.SelectedHex = &h
			} else {
				r.SelectedHex = nil
			}
		}
		if rl.IsMouseButtonPressed(rl.MouseButtonRight) {
			r.SelectedHex = nil
		}
	}
}

func (r *WorldRenderer) Draw(m *game.Map) {
	screenW := float32(rl.GetRenderWidth())
	screenH := float32(rl.GetRenderHeight())
	viewH := float32(r.viewport.Texture.Height)
	ratio := screenH / viewH
	viewW := float32(int32(screenW/ratio) + 1)

	if r.viewport.Texture.Width != int32(viewW) {
		rl.UnloadRenderTexture(r.viewport)
		r.viewport = rl.LoadRenderTexture(int32(viewW), int32(viewH))
	}
	r.Camera.Offset = rl.Vector2{
		X: float32(r.viewport.Texture.Width) / 2.0,
		Y: float32(r.viewport.Texture.Height) / 2.0,
	}

	srcRect, dstRect := r.viewportRects()
	rl.BeginTextureMode(r.viewport)
	r.drawBackground()

	mousePos := rlvec.FromRL(rl.GetScreenToWorld2D(rl.Vector2(r.MousePosition), r.Camera))
	rl.BeginMode2D(r.Camera)
	visible := r.drawMapTiles(m, mousePos)
	r.drawTileDetails(m, visible)
	r.drawBuildings(m, visible)
	r.drawTroops(m, visible)
	rl.EndMode2D()

	rl.EndTextureMode()
	rl.DrawTexturePro(r.viewport.Texture, srcRect, dstRect, rl.Vector2{}, 0.0, rl.White)
}

func (r *WorldRenderer) PixelToHex(position v.Vec2) game.Hex {
	q := (2.0 * position.X) / (3.0 * r.HexSize.X)
	axialR := (-position.X)/(3.0*r.HexSize.X) + position.Y/(sqrt3*r.HexSize.Y)
	return game.NewAxial(q, axialR).ToHex()
}

func (r *WorldRenderer) HexToPixel(hex v.Vec2i) v.Vec2 {
	x := hex.X
	y := hex.Y
	width := r.HexSize.X * 2.0
	height := r.HexSize.Y * sqrt3

	yOffset := height / 2.0 * float32(x%2)
	return v.Vec2{X: float32(x) * width / 4.0 * 3.0, Y: float32(y)*height + yOffset}
}

func (r *WorldRenderer) FocusOnHex(hex game.Hex) {
	pixelPos := r.HexToPixel(hex.Vec2i)
	r.TargetPosition = pixelPos
	r.InterpolateFocus = true
	r.PanVelocity = v.Vec2{}
}

func (r *WorldRenderer) viewportRects() (rl.Rectangle, rl.Rectangle) {
	screenW := float32(rl.GetRenderWidth())
	screenH := float32(rl.GetRenderHeight())
	viewH := float32(r.viewport.Texture.Height)
	ratio := screenH / viewH
	viewW := float32(int32(screenW/ratio) + 1)

	return rl.Rectangle{X: 0, Y: 0, Width: viewW, Height: -viewH}, rl.Rectangle{
		X:      (screenW - viewW*ratio) / 2.0,
		Y:      (screenH - viewH*ratio) / 2.0,
		Width:  viewW * ratio,
		Height: viewH * ratio,
	}
}

func (r *WorldRenderer) drawBackground() {
	if rl.IsShaderValid(shaders.WorldBackground) {
		rl.SetShaderValue(shaders.WorldBackground, shaders.WorldBackgroundTimeLoc, []float32{float32(rl.GetTime())}, rl.ShaderUniformFloat)
		rl.BeginShaderMode(shaders.WorldBackground)
		rl.Begin(rl.Triangles)

		width := float32(r.viewport.Texture.Width)
		height := float32(r.viewport.Texture.Height)
		rlutil.Color4ub(255, 255, 0, 255)
		rl.Normal3f(0.0, 0.0, 1.0)

		rl.TexCoord2f(0.0, 0.0)
		rlutil.Vertex2f(0, 0)
		rl.TexCoord2f(width, height)
		rlutil.Vertex2f(width, height)
		rl.TexCoord2f(width, 0.0)
		rlutil.Vertex2f(width, 0)
		rl.TexCoord2f(0.0, height)
		rlutil.Vertex2f(0, height)
		rl.TexCoord2f(width, height)
		rlutil.Vertex2f(width, height)
		rl.TexCoord2f(0.0, 0.0)
		rlutil.Vertex2f(0, 0)

		rl.End()
		rl.EndShaderMode()
	}

	if rl.IsShaderValid(shaders.Void) {
		rl.SetShaderValue(shaders.Void, shaders.VoidTimeLoc, []float32{float32(rl.GetTime())}, rl.ShaderUniformFloat)
	}
}

func (r *WorldRenderer) drawMapTiles(m *game.Map, mousePos v.Vec2) []visibleTile {
	topLeft := rl.GetScreenToWorld2D(rl.Vector2{}, r.Camera)
	topLeft = rl.Vector2Subtract(topLeft, rl.Vector2{X: r.HexSize.X * 2.0, Y: r.HexSize.Y * 2.0})
	bottomRight := rl.GetScreenToWorld2D(rl.Vector2{
		X: float32(r.viewport.Texture.Width),
		Y: float32(r.viewport.Texture.Height),
	}, r.Camera)
	bottomRight = rl.Vector2Add(bottomRight, rl.Vector2{X: r.HexSize.X * 2.0, Y: r.HexSize.Y * 2.0})

	width := r.HexSize.X * 2.0
	height := r.HexSize.Y * sqrt3
	var hoveredHex game.Hex
	hoverWorld := !global.UIBlocksWorldInput
	if hoverWorld {
		hoveredHex = r.PixelToHex(mousePos)
	}
	visible := make([]visibleTile, 0)

	rl.Begin(rl.Triangles)
	for x := range m.Grid {
		for y, cell := range m.Grid[x] {
			yOffset := height / 2.0 * float32(x%2)
			worldPos := v.Vec2{X: float32(x) * width / 4.0 * 3.0, Y: float32(y)*height + yOffset}
			if worldPos.X < topLeft.X || worldPos.X > bottomRight.X || worldPos.Y < topLeft.Y || worldPos.Y > bottomRight.Y {
				continue
			}

			hex := game.NewHex(int32(x), int32(y))
			color := tileColor(cell.Tile)
			if cell.Owner >= 0 && int(cell.Owner) < len(factionColors) {
				color = rl.ColorLerp(color, factionColors[cell.Owner], 0.4)
			}
			if hoverWorld && hex == hoveredHex {
				color = *util.ColorAdd(color, 30)
			}
			if x%2 == 1 {
				color = *util.ColorSub(color, 12)
			}
			if y%2 == 1 {
				color = *util.ColorSub(color, 6)
			}

			drawHexagonBuffered(worldPos.X, worldPos.Y, r.HexSize, color)
			visible = append(visible, visibleTile{hex: hex, position: worldPos, tile: cell.Tile})
		}
	}
	rl.End()

	return visible
}

func (r *WorldRenderer) clampCameraToMap(m *game.Map) {
	hexWidth := r.HexSize.X * 2.0
	hexHeight := r.HexSize.Y * sqrt3
	// Include the full outer hexagons in the map bounds.
	worldMinX := -r.HexSize.X
	worldMaxX := float32(m.GridSize.X-1)*hexWidth/4.0*3.0 + r.HexSize.X
	worldMinY := -hexHeight / 2.0
	worldMaxY := float32(m.GridSize.Y) * hexHeight
	// Convert the screen-space padding to world units at the current zoom.
	worldPadding := cameraBoundsPadding / r.Camera.Zoom
	worldMinX -= worldPadding
	worldMaxX += worldPadding
	worldMinY -= worldPadding
	worldMaxY += worldPadding

	// Camera.Target sits at Camera.Offset, so clamp using visible half-sizes.
	viewHalfWidth := float32(r.viewport.Texture.Width) / (2.0 * r.Camera.Zoom)
	viewHalfHeight := float32(r.viewport.Texture.Height) / (2.0 * r.Camera.Zoom)
	minTargetX := worldMinX + viewHalfWidth
	maxTargetX := worldMaxX - viewHalfWidth
	minTargetY := worldMinY + viewHalfHeight
	maxTargetY := worldMaxY - viewHalfHeight
	target := rlvec.FromRL(r.Camera.Target)
	previousTarget := target

	// Center axes where the viewport is larger than the map.
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

	r.Camera.Target = rlvec.ToRL(target)
	if stopMomentumAtBounds {
		if target.X != previousTarget.X {
			r.PanVelocity.X = 0.0
		}
		if target.Y != previousTarget.Y {
			r.PanVelocity.Y = 0.0
		}
	}
}
