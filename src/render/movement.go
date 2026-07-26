package render

import (
	"image/color"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/rlvec"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

const (
	troopAnimationStep                 = 140 * time.Millisecond
	troopTrailFade                     = 250 * time.Millisecond
	routeCancelHitRadiusPixels float32 = 8
)

type troopAnimation struct {
	event   game.MovementEvent
	elapsed time.Duration
}

func (r *WorldRenderer) StartMovementAnimations(events []game.MovementEvent) {
	for _, event := range events {
		if len(event.Path) < 2 {
			continue
		}
		r.troopAnimations = append(r.troopAnimations, troopAnimation{
			event: game.MovementEvent{
				Troop: event.Troop,
				Owner: event.Owner,
				Path:  append([]game.Hex(nil), event.Path...),
			},
		})
	}
}

func (r *WorldRenderer) updateTroopAnimations(delta time.Duration) {
	filtered := r.troopAnimations[:0]
	for i := range r.troopAnimations {
		r.troopAnimations[i].elapsed += delta
		duration := time.Duration(len(r.troopAnimations[i].event.Path)-1)*troopAnimationStep + troopTrailFade
		if r.troopAnimations[i].elapsed < duration {
			filtered = append(filtered, r.troopAnimations[i])
		}
	}
	r.troopAnimations = filtered
}

func (r *WorldRenderer) MovementAnimating() bool {
	for _, animation := range r.troopAnimations {
		moveDuration := time.Duration(len(animation.event.Path)-1) * troopAnimationStep
		if animation.elapsed <= moveDuration {
			return true
		}
	}
	return false
}

func (r *WorldRenderer) drawMovementRoutes(m *game.Map) {
	r.drawActionTargets(m)
	for _, order := range r.Orders {
		path, stops, ok := movementOrderRoute(m, r.LocalFaction, order)
		if !ok {
			continue
		}
		col := color.RGBA{R: 245, G: 245, B: 245, A: 100}
		width := float32(3)
		if r.SelectedKind == SelectionTroop &&
			r.SelectedHex != nil &&
			*r.SelectedHex == order.Current {
			col.A = 190
			width = 5
		}
		r.drawPathLine(path, col, width)
		stopColor := rl.Gold
		stopColor.A = 170
		if r.SelectedKind == SelectionTroop &&
			r.SelectedHex != nil &&
			*r.SelectedHex == order.Current {
			stopColor.A = 230
		}
		for _, stop := range stops {
			r.drawMovementStop(stop, stopColor, false)
		}
	}
	if len(r.PreviewPath) > 1 {
		r.drawPathLine(r.PreviewPath, color.RGBA{R: 255, G: 255, B: 255, A: 230}, 6)
		for _, stop := range r.PreviewStops {
			r.drawMovementStop(stop, rl.Gold, true)
		}
	}
}

func movementOrderRoute(
	m *game.Map,
	faction int8,
	order game.MovementOrder,
) ([]game.Hex, []game.Hex, bool) {
	path, ok := m.FindTroopPath(faction, order.Current, order.Destination)
	if !ok {
		return nil, nil, false
	}
	source := m.GetCell(order.Current)
	if source == nil {
		return path, nil, true
	}
	stops := m.MovementTurnStops(path, game.TroopMovementBudget(source.Troop))
	return path, stops, true
}

func (r *WorldRenderer) cancelQueuedMovementAt(
	m *game.Map,
	point vec.Vec2,
	click bool,
) bool {
	if !click {
		return false
	}
	order, ok := r.movementOrderNear(m, point)
	if !ok {
		return false
	}
	if r.OnCancelMovement != nil && !r.OnCancelMovement(order.Current) {
		return false
	}
	r.RemoveMovementOrder(order.Current)
	return true
}

func (r *WorldRenderer) movementOrderNear(
	m *game.Map,
	point vec.Vec2,
) (game.MovementOrder, bool) {
	hitRadius := r.zoomSafeSize(0, routeCancelHitRadiusPixels)
	maxDistanceSquared := hitRadius * hitRadius
	bestDistanceSquared := maxDistanceSquared
	var best game.MovementOrder
	found := false

	for _, order := range r.Orders {
		path, _, ok := movementOrderRoute(m, r.LocalFaction, order)
		if !ok {
			continue
		}
		for i := 1; i < len(path); i++ {
			from := r.HexToPixel(path[i-1].Vec2i)
			to := r.HexToPixel(path[i].Vec2i)
			distanceSquared := pointSegmentDistanceSquared(point, from, to)
			if distanceSquared <= bestDistanceSquared {
				bestDistanceSquared = distanceSquared
				best = order
				found = true
			}
		}
	}
	return best, found
}

func pointSegmentDistanceSquared(point, from, to vec.Vec2) float32 {
	segment := to.Sub(from)
	lengthSquared := segment.MagnitudeSqr()
	if lengthSquared == 0 {
		return point.Sub(from).MagnitudeSqr()
	}
	offset := point.Sub(from)
	projection := (offset.X*segment.X + offset.Y*segment.Y) / lengthSquared
	projection = max(float32(0), min(float32(1), projection))
	nearest := from.Add(segment.Mul(vec.Vec2{X: projection, Y: projection}))
	return point.Sub(nearest).MagnitudeSqr()
}

func (r *WorldRenderer) drawMovementStop(hex game.Hex, col color.RGBA, emphasized bool) {
	position := r.HexToPixel(hex.Vec2i)
	radius := float32(5)
	if emphasized {
		radius = 7
	}
	radius = r.zoomSafeSize(radius, 2.5)
	rl.DrawCircleV(rlvec.ToRL(position), radius, col)
	inner := color.RGBA{R: 28, G: 31, B: 36, A: col.A}
	rl.DrawCircleV(rlvec.ToRL(position), radius*0.45, inner)
}

func (r *WorldRenderer) drawActionTargets(m *game.Map) {
	if r.SelectedHex == nil {
		return
	}
	from := *r.SelectedHex
	source := m.GetCell(from)
	if source == nil {
		return
	}

	selectedPos := r.HexToPixel(from.Vec2i)
	rl.DrawCircleLinesV(rlvec.ToRL(selectedPos), 17, rl.White)

	for x := range m.Grid {
		for y := range m.Grid[x] {
			to := game.NewHex(int32(x), int32(y))
			if from != to && !game.HexAdjacent(from, to) {
				continue
			}
			valid := false
			col := rl.Green
			switch {
			case r.SelectedKind == SelectionTroop &&
				r.BuildingToPlace != game.BuildingUnknown:
				valid = r.canBuildAt(m, from, to, r.BuildingToPlace)
			case r.SelectedKind == SelectionBuilding &&
				r.RecruitToPlace != game.TroopUnknown:
				valid = r.canRecruitAt(m, from, to, r.RecruitToPlace)
			case r.SelectedKind == SelectionTroop &&
				source.Troop != game.TroopUnknown &&
				source.Troop != game.TroopScout &&
				source.TroopOwner == r.LocalFaction:
				target := m.GetCell(to)
				valid = target != nil &&
					target.Building != game.BuildingUnknown &&
					target.Building != game.BuildingTownhall &&
					target.Owner >= 0 &&
					target.Owner != r.LocalFaction
				col = rl.Red
			}
			if !valid {
				continue
			}
			position := r.HexToPixel(to.Vec2i)
			rl.DrawCircleLinesV(rlvec.ToRL(position), 19, col)
		}
	}
}

func (r *WorldRenderer) drawPathLine(path []game.Hex, col color.RGBA, width float32) {
	for i := 1; i < len(path); i++ {
		from := r.HexToPixel(path[i-1].Vec2i)
		to := r.HexToPixel(path[i].Vec2i)
		rl.DrawLineEx(rlvec.ToRL(from), rlvec.ToRL(to), r.zoomSafeSize(width, 0.9), col)
	}
}

func (r *WorldRenderer) zoomSafeSize(worldSize, minimumScreenPixels float32) float32 {
	if r.Camera.Zoom <= 0 {
		return worldSize
	}
	return max(worldSize, minimumScreenPixels/r.Camera.Zoom)
}

func (r *WorldRenderer) drawTroopAnimations() {
	for _, animation := range r.troopAnimations {
		event := animation.event
		moveDuration := time.Duration(len(event.Path)-1) * troopAnimationStep
		alpha := float32(1)
		if animation.elapsed > moveDuration {
			alpha = 1 - float32(animation.elapsed-moveDuration)/float32(troopTrailFade)
		}
		alpha = max(float32(0), min(float32(1), alpha))

		position, reached := r.animationPosition(animation)
		trail := append([]game.Hex(nil), event.Path[:reached+1]...)
		trailColor := factionColor(event.Owner)
		trailColor.A = uint8(float32(180) * alpha)
		r.drawPartialTrail(trail, position, trailColor)
		if animation.elapsed <= moveDuration {
			drawTroopMarker(position, r.HexSize, event.Troop, factionColor(event.Owner))
		}
	}
}

func (r *WorldRenderer) animationPosition(animation troopAnimation) (vec.Vec2, int) {
	path := animation.event.Path
	moveDuration := time.Duration(len(path)-1) * troopAnimationStep
	elapsed := min(animation.elapsed, moveDuration)
	segment := int(elapsed / troopAnimationStep)
	if segment >= len(path)-1 {
		return r.HexToPixel(path[len(path)-1].Vec2i), len(path) - 1
	}
	progress := float32(elapsed%troopAnimationStep) / float32(troopAnimationStep)
	from := r.HexToPixel(path[segment].Vec2i)
	to := r.HexToPixel(path[segment+1].Vec2i)
	return from.Lerp(to, progress), segment
}

func (r *WorldRenderer) drawPartialTrail(path []game.Hex, current vec.Vec2, col color.RGBA) {
	if len(path) == 0 {
		return
	}
	for i := 1; i < len(path); i++ {
		from := r.HexToPixel(path[i-1].Vec2i)
		to := r.HexToPixel(path[i].Vec2i)
		rl.DrawLineEx(rlvec.ToRL(from), rlvec.ToRL(to), r.zoomSafeSize(5, 0.9), col)
	}
	last := r.HexToPixel(path[len(path)-1].Vec2i)
	if last != current {
		rl.DrawLineEx(rlvec.ToRL(last), rlvec.ToRL(current), r.zoomSafeSize(5, 0.9), col)
	}
}

func factionColor(owner int8) color.RGBA {
	if owner < 0 || int(owner) >= len(factionColors) {
		return rl.White
	}
	return factionColors[owner]
}

func (r *WorldRenderer) troopEndpointAnimating(hex game.Hex, owner int8, troop game.TroopType) bool {
	for _, animation := range r.troopAnimations {
		path := animation.event.Path
		moveDuration := time.Duration(len(path)-1) * troopAnimationStep
		if animation.elapsed <= moveDuration &&
			animation.event.Owner == owner &&
			animation.event.Troop == troop &&
			path[len(path)-1] == hex {
			return true
		}
	}
	return false
}
