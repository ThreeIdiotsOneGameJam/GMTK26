package render

import (
	"math"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/global"
	"github.com/threeidiotsonegamejam/gmtk26/src/render/shaders"
	"github.com/threeidiotsonegamejam/gmtk26/src/settings"
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

	// A short left press remains a click. Holding for this long, or moving this
	// far first, turns the same gesture into a camera pan.
	leftPanHoldSeconds   float32 = 0.18
	leftPanDragThreshold float32 = 6.0

	// Careful, short drags use raw momentum instead of the speed compression
	// curve. Distance and speed are in viewport pixels, independent of zoom.
	panMomentumShortDragDistance float32 = 32.0
	panMomentumSlowDragSpeed     float32 = 220.0

	// Stationary right-clicks may cancel a pending building or queued route,
	// while any larger gesture remains pure map panning.
	rightActionClickDistance float32 = 4.0

	// Treat sub-pixel samples as a stationary mouse. While stationary, quickly
	// bleed off the sampled velocity and cancel it entirely after a short hold.
	panStationarySampleDistance  float32 = 0.75
	panStationaryVelocityDamping         = float32(20.0)
	panStationaryCancelDelay     float32 = 0.12

	// Smooth camera movement toward focus target.
	cameraFocusSmoothness float32 = 4.0

	// Zoom smoothness used when focusing on a hex (slower than wheel).
	cameraFocusZoomSmoothness float32 = 2.0
)

type WorldRenderer struct {
	Camera        rl.Camera2D
	ZoomAnchor    v.Vec2
	TargetZoom    float32
	HexSize       v.Vec2
	PanStart      v.Vec2
	PanVelocity   v.Vec2
	MousePosition v.Vec2

	panDragging           bool
	panDragDistance       float32
	panDragDuration       float32
	panStationaryDuration float32
	panRawVelocity        v.Vec2
	leftPressPos          v.Vec2
	leftPressTime         float32
	leftGesture           bool
	leftPanning           bool
	rightGesture          bool

	HoveredHex     game.Hex
	SelectedHex    *game.Hex
	SelectedKind   SelectionKind
	LocalFaction   int8
	LocalCoins     int32
	LocalResources game.Resources
	ActionsEnabled bool

	Orders        []game.MovementOrder
	AttackOrders  []game.AttackOrder
	PreviewPath   []game.Hex
	PreviewStops  []game.Hex
	PreviewTarget *game.Hex

	TargetPosition   v.Vec2
	InterpolateFocus bool
	zoomSmoothness   float32

	BuildingToPlace game.BuildingType
	RecruitToPlace  game.UnitType
	// OnPlaceBuilding, when set, may take over a placement click. Returning
	// true means the click was handled externally (e.g. sent to the server)
	// and the map is not modified locally.
	OnPlaceBuilding  func(from, to game.Hex, building game.BuildingType) bool
	OnRecruit        func(from, to game.Hex, unit game.UnitType) bool
	OnMove           func(from, to game.Hex) bool
	OnAttack         func(from, to game.Hex) bool
	OnCancelMovement func(from game.Hex) bool
	OnCancelBuilding func(to game.Hex) bool
	buildingPreview  buildingPreview
	queuedBuilding   buildingPreview

	buildingsTexture rl.Texture2D
	unitsTexture     rl.Texture2D

	viewport    rl.RenderTexture2D
	initialized bool

	unitAnimations   []unitAnimation
	attackAnimations []attackAnimation
	selectionMenu    selectionMenu
}

func (r *WorldRenderer) Init(m *game.Map) {
	if r.initialized {
		return
	}
	r.initialized = true

	r.buildingsTexture = rl.LoadTexture("assets/textures/buildings.png")
	r.unitsTexture = rl.LoadTexture("assets/textures/units.png")

	if r.HexSize == (v.Vec2{}) {
		r.HexSize = v.Vec2{X: 48.0, Y: 48.0}
	}

	// A game-start packet can request Town Hall focus before the world's first
	// draw initializes the renderer. Preserve that target across camera setup.
	focusPending := r.InterpolateFocus
	focusTarget := r.TargetPosition
	focusZoom := r.TargetZoom
	focusZoomSmoothness := r.zoomSmoothness
	r.ResetCamera(m)
	if focusPending {
		r.TargetPosition = focusTarget
		r.TargetZoom = focusZoom
		r.zoomSmoothness = focusZoomSmoothness
		r.InterpolateFocus = true
	}

	r.viewport = rl.LoadRenderTexture(global.ViewportSize.X, global.ViewportSize.Y)
}

func (r *WorldRenderer) ResetCamera(m *game.Map) {
	if r.HexSize == (v.Vec2{}) {
		r.HexSize = v.Vec2{X: 48.0, Y: 48.0}
	}

	r.Camera.Zoom = cameraDefaultZoom
	r.TargetZoom = cameraDefaultZoom
	r.PanVelocity = v.Vec2{}
	r.panDragging = false
	r.panDragDistance = 0.0
	r.panDragDuration = 0.0
	r.panStationaryDuration = 0.0
	r.panRawVelocity = v.Vec2{}
	r.leftPressTime = 0.0
	r.leftGesture = false
	r.leftPanning = false
	r.rightGesture = false
	r.InterpolateFocus = false
	r.Camera.Target = rlvec.ToRL(v.Vec2{
		X: float32(m.GridSize.X),
		Y: float32(m.GridSize.Y),
	}.Mul(r.HexSize).Sub(global.ViewportSize.Vec2()))
	r.Camera.Offset = rlvec.ToRL(global.ViewportSize.Vec2().Mul(v.Vec2{X: 0.5, Y: 0.5}))
	r.TargetPosition = rlvec.FromRL(r.Camera.Target)
	r.zoomSmoothness = cameraZoomSmoothness
}

func (r *WorldRenderer) Unload() {
	if !r.initialized {
		return
	}
	rl.UnloadRenderTexture(r.viewport)
	rl.UnloadTexture(r.buildingsTexture)
	rl.UnloadTexture(r.unitsTexture)
	r.initialized = false
}

func (r *WorldRenderer) Update(m *game.Map, delta time.Duration) {
	r.updateUnitAnimations(delta)
	r.updateAttackAnimations(delta)
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

	if !global.UIModalBlocksInput && rl.IsMouseButtonPressed(rl.MouseButtonMiddle) {
		r.clearMouseSlot()
	}

	worldInputBlocked := global.UIBlocksWorldInput
	leftPressed := rl.IsMouseButtonPressed(rl.MouseButtonLeft)
	leftDown := rl.IsMouseButtonDown(rl.MouseButtonLeft)
	worldLeftClick, leftPanDown, leftPanReleased := r.updateLeftGesture(
		deltaSeconds,
		leftPressed,
		leftDown,
		worldInputBlocked,
	)
	rightDown := rl.IsMouseButtonDown(rl.MouseButtonRight)
	panButtonDown := rightDown || leftPanDown

	if worldInputBlocked {
		r.clampCameraToMap(m)
	} else {
		if rl.IsMouseButtonPressed(rl.MouseButtonRight) {
			r.rightGesture = true
			r.beginPan(r.MousePosition)
		}

		if r.panDragging && panButtonDown {
			r.updatePan(deltaSeconds)
		}
	}

	if !global.UIModalBlocksInput {
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
		r.applyKeyboardMovement(moveDir, deltaSeconds)
		r.applyWheelZoom(rl.GetMouseWheelMove())
	}

	rightPanReleased := rl.IsMouseButtonReleased(rl.MouseButtonRight)
	rightActionClick := rightPanReleased &&
		r.rightGesture &&
		!worldInputBlocked &&
		!leftPanDown &&
		r.panDragDistance <= rightActionClickDistance
	if rightPanReleased {
		r.rightGesture = false
	}
	if r.panDragging && !panButtonDown && (rightPanReleased || leftPanReleased) {
		r.finishPan()
	}

	if !panButtonDown {
		r.Camera.Target = rlvec.ToRL(rlvec.FromRL(r.Camera.Target).Add(r.PanVelocity.Mul(v.Vec2{
			X: deltaSeconds,
			Y: deltaSeconds,
		})))

		panDecay := float32(math.Exp(float64(-panMomentumDamping * deltaSeconds)))
		r.PanVelocity = r.PanVelocity.Mul(v.Vec2{X: panDecay, Y: panDecay})
	}

	r.TargetZoom = max(cameraMinZoom, min(r.TargetZoom, cameraMaxZoom))
	zoomBlend := 1.0 - float32(math.Exp(float64(-r.zoomSmoothness*deltaSeconds)))
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
		// A hex near the map edge may not be centerable at the current zoom.
		// Interpolate toward the closest legal target so focus can settle
		// instead of fighting the camera clamp forever.
		focusTarget := r.clampCameraTargetToMap(m, r.TargetPosition)
		diff := focusTarget.Sub(currentTarget)
		if diff.MagnitudeSqr() > 0.01 {
			focusBlend := 1.0 - float32(math.Exp(float64(-cameraFocusSmoothness*deltaSeconds)))
			r.Camera.Target = rlvec.ToRL(currentTarget.Lerp(focusTarget, focusBlend))
		} else {
			r.Camera.Target = rlvec.ToRL(focusTarget)
			r.InterpolateFocus = false
		}
	}

	r.clampCameraToMap(m)

	mousePos := rlvec.FromRL(rl.GetScreenToWorld2D(rl.Vector2(r.MousePosition), r.Camera))
	hex := r.PixelToHex(mousePos)
	r.HoveredHex = hex
	cancelledBuilding := r.cancelQueuedBuildingAt(hex, rightActionClick)
	if !cancelledBuilding {
		r.cancelQueuedMovementAt(m, mousePos, rightActionClick)
	}
	placementConsumedClick := r.updateBuildingPlacement(m, hex, worldLeftClick)
	recruitmentConsumedClick := r.updateRecruitPlacement(m, hex, worldLeftClick)
	r.updateMovementPreview(m, hex)
	if !global.UIBlocksWorldInput &&
		!placementConsumedClick &&
		!recruitmentConsumedClick &&
		r.BuildingToPlace == game.BuildingUnknown &&
		r.RecruitToPlace == game.UnitUnknown {
		if worldLeftClick {
			r.handleWorldClick(m, hex)
		}
	}
}

func (r *WorldRenderer) applyKeyboardMovement(moveDir v.Vec2, deltaSeconds float32) {
	if moveDir.X == 0.0 && moveDir.Y == 0.0 {
		return
	}

	zoomMovementScale := float32(math.Pow(float64(r.Camera.Zoom), cameraMoveZoomExponent))
	moveDistance := cameraMoveSpeed * deltaSeconds / zoomMovementScale
	r.Camera.Target = rlvec.ToRL(rlvec.FromRL(r.Camera.Target).Add(moveDir.Normalize().Mul(v.Vec2{
		X: moveDistance,
		Y: moveDistance,
	})))
	r.InterpolateFocus = false
}

func (r *WorldRenderer) applyWheelZoom(wheel float32) {
	if wheel == 0.0 {
		return
	}

	r.ZoomAnchor = r.MousePosition
	r.TargetZoom *= float32(math.Exp(float64(wheel * cameraZoomStep)))
	r.zoomSmoothness = cameraZoomSmoothness
	// Manual zoom takes ownership of the camera just like panning or WASD, so
	// an old focus target cannot pull cursor-anchored zooms back toward the
	// focused hex. Non-modal UI may still block clicks while allowing this.
	r.InterpolateFocus = false
}

func (r *WorldRenderer) updateLeftGesture(
	deltaSeconds float32,
	pressed bool,
	down bool,
	worldInputBlocked bool,
) (clicked bool, panning bool, panReleased bool) {
	if pressed {
		r.leftGesture = !worldInputBlocked
		r.leftPanning = false
		r.leftPressTime = 0.0
		r.leftPressPos = r.MousePosition
		if r.leftGesture {
			r.PanVelocity = v.Vec2{}
			r.panRawVelocity = v.Vec2{}
			r.InterpolateFocus = false
		}
	}

	if !r.leftGesture {
		return false, false, false
	}

	if down {
		r.leftPressTime += deltaSeconds
		dragDistance := r.MousePosition.Sub(r.leftPressPos)
		if !r.leftPanning &&
			(r.leftPressTime >= leftPanHoldSeconds ||
				dragDistance.MagnitudeSqr() >= leftPanDragThreshold*leftPanDragThreshold) {
			r.leftPanning = true
			r.beginPan(r.leftPressPos)
		}
		return false, r.leftPanning, false
	}

	clicked = !r.leftPanning && !worldInputBlocked
	panReleased = r.leftPanning
	r.leftGesture = false
	r.leftPanning = false
	return clicked, false, panReleased
}

func (r *WorldRenderer) beginPan(start v.Vec2) {
	r.PanStart = start
	r.PanVelocity = v.Vec2{}
	r.panDragging = true
	r.panDragDistance = 0.0
	r.panDragDuration = 0.0
	r.panStationaryDuration = 0.0
	r.panRawVelocity = v.Vec2{}
	r.InterpolateFocus = false
}

func (r *WorldRenderer) updatePan(deltaSeconds float32) {
	mouseDelta := r.MousePosition.Sub(r.PanStart)
	mouseDistance := mouseDelta.Magnitude()
	r.panDragDistance += mouseDistance
	r.panDragDuration += deltaSeconds
	panDelta := mouseDelta.Mul(v.Vec2{
		X: -1.0 / r.Camera.Zoom,
		Y: -1.0 / r.Camera.Zoom,
	})

	r.Camera.Target = rlvec.ToRL(rlvec.FromRL(r.Camera.Target).Add(panDelta))

	if mouseDistance <= panStationarySampleDistance {
		r.panStationaryDuration += deltaSeconds
		stationaryDecay := float32(math.Exp(float64(-panStationaryVelocityDamping * deltaSeconds)))
		r.PanVelocity = r.PanVelocity.Mul(v.Vec2{X: stationaryDecay, Y: stationaryDecay})
		r.panRawVelocity = r.panRawVelocity.Mul(v.Vec2{X: stationaryDecay, Y: stationaryDecay})
	} else {
		r.panStationaryDuration = 0.0
	}

	if deltaSeconds > 0.0 && mouseDistance > panStationarySampleDistance {
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
			r.panRawVelocity = r.panRawVelocity.Mul(v.Vec2{
				X: previousVelocityWeight,
				Y: previousVelocityWeight,
			}).Add(rawVelocity.Mul(v.Vec2{
				X: panVelocitySampleWeight,
				Y: panVelocitySampleWeight,
			}))
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

func (r *WorldRenderer) finishPan() {
	r.PanVelocity = panMomentumReleaseVelocity(
		r.PanVelocity,
		r.panRawVelocity,
		r.panDragDistance,
		r.panDragDuration,
		r.panStationaryDuration,
	)
	r.panDragging = false
}

func (r *WorldRenderer) handleWorldClick(m *game.Map, hex game.Hex) {
	if r.MovementAnimating() {
		return
	}
	cell := m.GetCell(hex)
	if cell == nil {
		r.clearSelection()
		return
	}

	if r.ActionsEnabled && r.SelectedHex != nil && r.SelectedKind == SelectionUnit {
		from := *r.SelectedHex
		source := m.GetCell(from)
		if source != nil && source.HasUnits() &&
			source.Units[0].Owner == r.LocalFaction {
			if hex == from {
				if r.hasMovementOrder(from) && r.OnCancelMovement != nil {
					if r.OnCancelMovement(from) {
						r.RemoveMovementOrder(from)
					}
				} else {
					r.clearSelection()
				}
				return
			}
			if cell.HasUnits() && cell.Units[0].Owner == r.LocalFaction {
				r.selectCell(hex, cell)
				return
			}
			if source.Units[0].Type != game.UnitScout &&
				(hasEnemyUnit(cell, r.LocalFaction) || hasEnemyBuilding(cell, r.LocalFaction)) {
				if r.OnAttack != nil && r.OnAttack(from, hex) {
					r.ClearQueuedBuilding()
					r.clearSelection()
				}
				return
			}
			if _, ok := m.FindUnitPath(r.LocalFaction, from, hex); ok {
				if r.OnMove != nil && r.OnMove(from, hex) {
					r.QueueMovementOrder(from, hex)
					r.clearSelection()
				}
				return
			}
		}
	}

	if cell.HasBuilding() || cell.HasUnits() {
		r.selectCell(hex, cell)
		return
	}
	r.clearSelection()
}

func (r *WorldRenderer) updateMovementPreview(m *game.Map, hovered game.Hex) {
	r.PreviewPath = nil
	r.PreviewStops = nil
	r.PreviewTarget = nil
	if !r.ActionsEnabled ||
		r.BuildingToPlace != game.BuildingUnknown ||
		r.RecruitToPlace != game.UnitUnknown ||
		r.SelectedHex == nil ||
		r.SelectedKind != SelectionUnit ||
		hovered == *r.SelectedHex {
		return
	}
	source := m.GetCell(*r.SelectedHex)
	if source == nil || !source.HasUnits() || source.Units[0].Owner != r.LocalFaction {
		return
	}
	path, ok := m.FindUnitPath(r.LocalFaction, *r.SelectedHex, hovered)
	if !ok {
		if hasEnemyUnit(m.GetCell(hovered), r.LocalFaction) ||
			hasEnemyBuilding(m.GetCell(hovered), r.LocalFaction) {
			approachPath, _, approachOk := m.FindAdjacentApproachPath(
				r.LocalFaction, *r.SelectedHex, hovered,
			)
			if approachOk {
				r.PreviewPath = approachPath
				r.PreviewTarget = &hovered
				r.PreviewStops = m.MovementTurnStops(approachPath, game.UnitMovementBudget(source.Units[0].Type))
			}
		}
		return
	}
	r.PreviewPath = path
	r.PreviewStops = m.MovementTurnStops(path, game.UnitMovementBudget(source.Units[0].Type))
}

func (r *WorldRenderer) SetMovementOrders(orders []game.MovementOrder) {
	r.Orders = append(r.Orders[:0], orders...)
}

func (r *WorldRenderer) SetAttackOrders(orders []game.AttackOrder) {
	r.AttackOrders = append(r.AttackOrders[:0], orders...)
}

func (r *WorldRenderer) QueueMovementOrder(from, destination game.Hex) {
	r.RemoveMovementOrder(from)
	r.Orders = append(r.Orders, game.MovementOrder{Current: from, Destination: destination})
}

func (r *WorldRenderer) RemoveMovementOrder(from game.Hex) {
	filtered := r.Orders[:0]
	for _, order := range r.Orders {
		if order.Current != from {
			filtered = append(filtered, order)
		}
	}
	r.Orders = filtered
}

func (r *WorldRenderer) hasMovementOrder(from game.Hex) bool {
	for _, order := range r.Orders {
		if order.Current == from {
			return true
		}
	}
	return false
}

func panMomentumReleaseVelocity(
	compressedVelocity,
	rawVelocity v.Vec2,
	distance,
	duration,
	stationaryDuration float32,
) v.Vec2 {
	if stationaryDuration >= panStationaryCancelDelay {
		return v.Vec2{}
	}
	if duration > 0.0 &&
		distance <= panMomentumShortDragDistance &&
		distance/duration <= panMomentumSlowDragSpeed {
		return rawVelocity
	}
	return compressedVelocity
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
	r.drawMovementRoutes(m)
	r.drawBuildings(m, visible, mousePos)
	r.drawUnits(m, visible)
	r.drawUnitAnimations()
	r.drawAttackAnimations()
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

// HexScreenPosition projects a world tile center into window UI coordinates.
// Keeping this conversion in the renderer ensures world-anchored UI follows
// camera panning and zooming while still rendering at a fixed UI scale.
func (r *WorldRenderer) HexScreenPosition(hex game.Hex) v.Vec2 {
	world := r.HexToPixel(hex.Vec2i)
	viewport := v.Vec2{
		X: (world.X-r.Camera.Target.X)*r.Camera.Zoom + r.Camera.Offset.X,
		Y: (world.Y-r.Camera.Target.Y)*r.Camera.Zoom + r.Camera.Offset.Y,
	}
	src, dst := r.viewportRects()
	if src.Width == 0 || src.Height == 0 {
		return viewport
	}
	return v.Vec2{
		X: dst.X + viewport.X*(dst.Width/src.Width),
		Y: dst.Y + viewport.Y*(dst.Height/-src.Height),
	}
}

func (r *WorldRenderer) FocusOnHex(hex game.Hex) {
	pixelPos := r.HexToPixel(hex.Vec2i)
	r.TargetPosition = pixelPos
	r.TargetZoom = cameraDefaultZoom
	r.zoomSmoothness = cameraFocusZoomSmoothness
	// Focus zooms around the viewport center. Cursor anchoring only applies
	// after the player takes over with the mouse wheel.
	r.ZoomAnchor = rlvec.FromRL(r.Camera.Offset)
	r.InterpolateFocus = true
	r.PanVelocity = v.Vec2{}
	r.panRawVelocity = v.Vec2{}
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
		if !settings.Current.ReducedMotion {
			rl.SetShaderValue(shaders.WorldBackground, shaders.WorldBackgroundTimeLoc, []float32{float32(rl.GetTime())}, rl.ShaderUniformFloat)
		}
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

	batch := newHexBatch()
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
				color = lerpColor(color, factionColors[cell.Owner], 0.4)
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

			batch.Add(worldPos.X, worldPos.Y, r.HexSize, color)
			visible = append(visible, visibleTile{hex: hex, position: worldPos, tile: cell.Tile})
		}
	}
	batch.Draw()

	return visible
}

func (r *WorldRenderer) clampCameraToMap(m *game.Map) {
	target := rlvec.FromRL(r.Camera.Target)
	previousTarget := target
	target = r.clampCameraTargetToMap(m, target)

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

func (r *WorldRenderer) clampCameraTargetToMap(m *game.Map, target v.Vec2) v.Vec2 {
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

	return target
}
