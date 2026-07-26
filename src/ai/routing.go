package ai

import (
	"sort"

	"github.com/threeidiotsonegamejam/gmtk26/src/game"
)

type routeKind uint8

const (
	routeMove routeKind = iota
	routeAttack
)

type pathKey struct {
	kind routeKind
	from game.Hex
	to   game.Hex
}

type pathResult struct {
	ok       bool
	approach game.Hex
}

type pathOracle struct {
	world   *WorldSnapshot
	owner   int8
	limit   int
	queries int
	cache   map[pathKey]pathResult
}

func newPathOracle(world *WorldSnapshot, owner int8, limit int) *pathOracle {
	return &pathOracle{
		world: world,
		owner: owner,
		limit: limit,
		cache: make(map[pathKey]pathResult),
	}
}

func (oracle *pathOracle) canMove(from, to game.Hex) bool {
	key := pathKey{kind: routeMove, from: from, to: to}
	if result, exists := oracle.cache[key]; exists {
		return result.ok
	}
	if oracle.queries >= oracle.limit {
		return false
	}
	oracle.queries++
	_, ok := oracle.world.Map.FindUnitPath(oracle.owner, from, to)
	oracle.cache[key] = pathResult{ok: ok}
	return ok
}

func (oracle *pathOracle) attackApproach(from, to game.Hex) (game.Hex, bool) {
	key := pathKey{kind: routeAttack, from: from, to: to}
	if result, exists := oracle.cache[key]; exists {
		return result.approach, result.ok
	}
	if oracle.queries >= oracle.limit {
		return game.Hex{}, false
	}
	oracle.queries++
	_, approach, ok := oracle.world.Map.FindAdjacentApproachPath(oracle.owner, from, to)
	oracle.cache[key] = pathResult{ok: ok, approach: approach}
	return approach, ok
}

type routeProposal struct {
	command OrderCommand
	utility float64
}

func planRoutes(
	world *WorldSnapshot,
	own OwnState,
	analysis worldAnalysis,
	goal Goal,
	target targetRef,
	hasTarget bool,
	oracle *pathOracle,
	config Config,
) ([]OrderCommand, float64) {
	units := append([]unitRef(nil), analysis.ownUnits...)
	sort.SliceStable(units, func(i, j int) bool {
		if units[i].Position.X == units[j].Position.X {
			return units[i].Position.Y < units[j].Position.Y
		}
		return units[i].Position.X < units[j].Position.X
	})

	moves := indexMovementOrders(own.MovementOrders)
	attacks := indexAttackOrders(own.AttackOrders)
	reserved := make(map[game.Hex]bool)
	proposals := make([]routeProposal, 0, len(units))

	for _, unit := range units {
		if existingRouteStillUseful(
			world,
			unit,
			moves,
			attacks,
			analysis,
			goal,
			target,
			hasTarget,
			config,
		) {
			if order, ok := moves[unit.Position]; ok {
				reserved[order.Destination] = true
			}
			continue
		}

		if unit.Data.Type == game.UnitScout {
			proposal, ok := scoutRoute(unit, analysis, reserved, oracle)
			if ok {
				reserved[proposal.command.To] = true
				proposals = append(proposals, proposal)
			}
			continue
		}

		combatTarget, ok := selectCombatTarget(unit, analysis, goal, target, hasTarget)
		if !ok || game.HexAdjacent(unit.Position, combatTarget.Position) {
			continue
		}
		approach, reachable := oracle.attackApproach(unit.Position, combatTarget.Position)
		if !reachable || reserved[approach] {
			continue
		}
		reserved[approach] = true
		proposals = append(proposals, routeProposal{
			command: OrderCommand{
				Kind: OrderAttack,
				From: unit.Position,
				To:   combatTarget.Position,
			},
			utility: combatTarget.Score,
		})
	}

	sort.SliceStable(proposals, func(i, j int) bool {
		if proposals[i].utility == proposals[j].utility {
			if proposals[i].command.From.X == proposals[j].command.From.X {
				return proposals[i].command.From.Y < proposals[j].command.From.Y
			}
			return proposals[i].command.From.X < proposals[j].command.From.X
		}
		return proposals[i].utility > proposals[j].utility
	})
	commands := make([]OrderCommand, len(proposals))
	best := analysis.signals.routeValue
	for i, proposal := range proposals {
		commands[i] = proposal.command
		if proposal.utility > best {
			best = proposal.utility
		}
	}
	return commands, clamp01(best)
}

func existingRouteStillUseful(
	world *WorldSnapshot,
	unit unitRef,
	moves map[game.Hex]game.MovementOrder,
	attacks map[game.Hex]game.AttackOrder,
	analysis worldAnalysis,
	goal Goal,
	strategic targetRef,
	hasStrategic bool,
	config Config,
) bool {
	if order, ok := attacks[unit.Position]; ok {
		target := world.Map.GetCell(order.TargetTile)
		if target == nil || !cellHasEnemy(target, analysis.owner) {
			return false
		}
		if goal == GoalDefend &&
			analysis.signals.townhallThreat >= config.EmergencyThreat &&
			analysis.hasTownhall &&
			order.TargetTile.Distance(analysis.townhall) > 10 {
			return false
		}
		if hasStrategic && strategic.Position != order.TargetTile {
			currentScore := targetUtilityAt(analysis.targets, order.TargetTile)
			if strategic.Score > max(currentScore, 0.01)*config.TargetSwitchMultiplier {
				return false
			}
		}
		return true
	}
	if order, ok := moves[unit.Position]; ok {
		target := world.Map.GetCell(order.Destination)
		if target == nil || target.HasUnits() && order.Destination != unit.Position {
			return false
		}
		if unit.Data.Type != game.UnitScout &&
			goal == GoalDefend &&
			analysis.signals.townhallThreat >= config.EmergencyThreat {
			return false
		}
		if unit.Data.Type == game.UnitScout && len(analysis.sites) > 0 {
			currentScore := siteUtilityAt(analysis.sites, order.Destination)
			if analysis.sites[0].Position != order.Destination &&
				analysis.sites[0].Score > max(currentScore, 0.01)*config.TargetSwitchMultiplier {
				return false
			}
		}
		return true
	}
	return false
}

func targetUtilityAt(targets []targetRef, position game.Hex) float64 {
	for _, target := range targets {
		if target.Position == position {
			return target.Score
		}
	}
	return 0
}

func siteUtilityAt(sites []siteRef, position game.Hex) float64 {
	best := 0.0
	for _, site := range sites {
		if site.Position == position && site.Score > best {
			best = site.Score
		}
	}
	return best
}

func scoutRoute(
	unit unitRef,
	analysis worldAnalysis,
	reserved map[game.Hex]bool,
	oracle *pathOracle,
) (routeProposal, bool) {
	limit := min(16, len(analysis.sites))
	for _, site := range analysis.sites[:limit] {
		if reserved[site.Position] ||
			unit.Position == site.Position ||
			game.HexAdjacent(unit.Position, site.Position) {
			continue
		}
		if !oracle.canMove(unit.Position, site.Position) {
			continue
		}
		return routeProposal{
			command: OrderCommand{
				Kind: OrderMove,
				From: unit.Position,
				To:   site.Position,
			},
			utility: site.Score,
		}, true
	}
	return routeProposal{}, false
}

func selectCombatTarget(
	unit unitRef,
	analysis worldAnalysis,
	goal Goal,
	strategic targetRef,
	hasStrategic bool,
) (targetRef, bool) {
	if goal == GoalDefend && analysis.hasTownhall {
		bestDistance := int32(1<<31 - 1)
		var best targetRef
		found := false
		for _, enemy := range analysis.enemyUnits {
			distance := enemy.Position.Distance(analysis.townhall)
			if distance <= 10 && distance < bestDistance {
				bestDistance = distance
				best = targetRef{
					Position: enemy.Position,
					Owner:    enemy.Data.Owner,
					Score:    clamp01(1 - float64(distance)/12),
				}
				found = true
			}
		}
		if found {
			return best, true
		}
	}
	if hasStrategic {
		return strategic, true
	}
	if len(analysis.targets) == 0 {
		return targetRef{}, false
	}
	best := analysis.targets[0]
	for _, candidate := range analysis.targets[1:min(8, len(analysis.targets))] {
		if unit.Position.Distance(candidate.Position) < unit.Position.Distance(best.Position) &&
			candidate.Score >= best.Score*0.75 {
			best = candidate
		}
	}
	return best, true
}

func indexMovementOrders(orders []game.MovementOrder) map[game.Hex]game.MovementOrder {
	indexed := make(map[game.Hex]game.MovementOrder, len(orders))
	for _, order := range orders {
		indexed[order.Current] = order
	}
	return indexed
}

func indexAttackOrders(orders []game.AttackOrder) map[game.Hex]game.AttackOrder {
	indexed := make(map[game.Hex]game.AttackOrder, len(orders))
	for _, order := range orders {
		indexed[order.From] = order
	}
	return indexed
}

func cellHasEnemy(cell *game.Cell, owner int8) bool {
	if cell == nil {
		return false
	}
	for _, unit := range cell.Units {
		if unit.Owner != owner {
			return true
		}
	}
	return cell.HasBuilding() && cell.Owner >= 0 && cell.Owner != owner
}
