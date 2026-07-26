package server

import (
	"fmt"

	"github.com/threeidiotsonegamejam/gmtk26/src/game"
)

type roundIntent struct {
	faction     int
	action      game.ActionType
	automatic   bool
	from        game.Hex
	to          game.Hex
	destination game.Hex
	unit        game.UnitType
	building    game.BuildingType
	path        []game.Hex
	cost        int32
	targetsTile bool
	valid       bool
	result      game.ActionResult
}

func (gi *GameInstance) processClientActions() {
	gi.actionResults = make(map[int]*game.ActionResult)
	gi.movementEvents = nil
	gi.attackEvents = nil

	// Phase 1: auto-attacks from persistent AttackOrders (no contest)
	for factionIdx := range gi.game.Factions {
		allOrders := gi.attackOrders[factionIdx]
		remaining := allOrders[:0]
		for _, order := range allOrders {
			source := gi.game.Map.GetCell(order.From)
			target := gi.game.Map.GetCell(order.TargetTile)
			if source == nil || target == nil || !source.HasUnits() ||
				source.Units[0].Owner != int8(factionIdx) {
				continue
			}
			if !game.HexAdjacent(order.From, order.TargetTile) {
				remaining = append(remaining, order)
				continue
			}
			if !gi.resolveAttack(factionIdx, order.From, order.TargetTile) {
				continue
			}
			if hasEnemies(target, int8(factionIdx)) {
				remaining = append(remaining, order)
			}
		}
		gi.attackOrders[factionIdx] = remaining
	}

	// Phase 2: manual actions + auto-movement
	intents := make([]*roundIntent, len(gi.game.Factions))
	for factionIdx := range gi.game.Factions {
		action, submitted := gi.actions[factionIdx]
		if submitted && action != nil {
			intents[factionIdx] = gi.planManualIntent(factionIdx, action)
		} else {
			intents[factionIdx] = gi.planAutomaticMovement(factionIdx)
		}
	}

	contested := make(map[game.Hex]bool)
	targeters := make(map[game.Hex][]*roundIntent)
	for _, intent := range intents {
		if intent != nil && intent.valid && intent.targetsTile {
			targeters[intent.to] = append(targeters[intent.to], intent)
		}
	}
	for target, contenders := range targeters {
		if len(contenders) > 1 {
			contested[target] = true
		}
	}

	for _, intent := range intents {
		if intent == nil {
			continue
		}
		if !intent.valid {
			gi.recordResult(intent)
			continue
		}
		if contested[intent.to] && intent.targetsTile {
			intent.result.Status = game.ActionResultContested
			intent.result.Message = "Action cancelled: destination was contested"
			if intent.action == game.ActionMove {
				gi.movementOrders[intent.faction] = removeMovementOrder(
					gi.movementOrders[intent.faction], intent.from,
				)
			}
			gi.recordResult(intent)
			continue
		}
		gi.applyIntent(intent)
		gi.recordResult(intent)
	}
	gi.movementPriorities = make(map[int]game.Hex)
}

func hasEnemies(cell *game.Cell, factionOwner int8) bool {
	for _, u := range cell.Units {
		if u.Owner != factionOwner {
			return true
		}
	}
	return cell.HasBuilding() && cell.Owner >= 0 && cell.Owner != factionOwner
}

func (gi *GameInstance) planManualIntent(factionIdx int, action *submittedAction) *roundIntent {
	intent := gi.newIntent(factionIdx, action.Type, false)
	faction := &gi.game.Factions[factionIdx]
	factionOwner := int8(factionIdx)
	m := &gi.game.Map

	switch action.Type {
	case game.ActionPass:
		intent.valid = true
		intent.result.Status = game.ActionResultSucceeded
		intent.result.Message = "Holding position this round"

	case game.ActionRecruit:
		if action.Recruit == nil {
			return intent.invalid("Recruit payload was missing")
		}
		intent.from, intent.to, intent.unit = action.Recruit.From, action.Recruit.To, action.Recruit.Unit
		source, target := m.GetCell(intent.from), m.GetCell(intent.to)
		if source == nil || target == nil || source.Owner != factionOwner {
			return intent.invalid("Recruitment source is not friendly")
		}
		if !game.HexAdjacent(intent.from, intent.to) {
			return intent.invalid("Recruitment target must be adjacent")
		}
		if game.TerrainMovementCost(target.Tile) <= 0 ||
			(target.Owner != -1 && target.Owner != factionOwner) ||
			target.HasUnits() {
			return intent.invalid("Recruitment target is not available")
		}
		validUnit := intent.unit >= game.UnitPeasant && intent.unit <= game.UnitScout
		if !validUnit ||
			source.BuildingType() == game.BuildingTownhall && intent.unit != game.UnitScout ||
			source.BuildingType() != game.BuildingTownhall && source.BuildingType() != game.BuildingBarracks {
			return intent.invalid("Building cannot recruit that unit")
		}
		intent.cost = game.UnitCost(intent.unit)
		if faction.Coins < intent.cost {
			return intent.insufficientCoins()
		}
		for res, amt := range game.UnitResourceCost(intent.unit) {
			if faction.Resources[res] < amt {
				return intent.insufficientResources(res)
			}
		}
		intent.targetsTile = true
		intent.valid = true
		intent.result.Status = game.ActionResultSucceeded
		intent.result.Message = fmt.Sprintf("%s recruited", intent.unit)

	case game.ActionBuild:
		if action.Build == nil {
			return intent.invalid("Build payload was missing")
		}
		intent.from, intent.to, intent.building = action.Build.From, action.Build.To, action.Build.Building
		source, target := m.GetCell(intent.from), m.GetCell(intent.to)
		if source == nil || target == nil ||
			!source.HasUnits() ||
			source.Units[0].Type != game.UnitScout ||
			source.Units[0].Owner != factionOwner {
			return intent.invalid("A friendly Scout is required to build")
		}
		if intent.from != intent.to && !game.HexAdjacent(intent.from, intent.to) {
			return intent.invalid("Scout may only build on its tile or an adjacent tile")
		}
		if target.Owner != -1 && target.Owner != factionOwner {
			return intent.invalid("Cannot build in enemy territory")
		}
		if target.HasUnits() && target.Units[0].Owner != factionOwner {
			return intent.invalid("Cannot build beneath an enemy unit")
		}
		if !game.BuildingCanPlace(m, intent.building, intent.to) {
			return intent.invalid("Building cannot be placed on that terrain")
		}
		intent.cost = game.BuildingCost(intent.building)
		if faction.Coins < intent.cost {
			return intent.insufficientCoins()
		}
		intent.targetsTile = true
		intent.valid = true
		intent.result.Status = game.ActionResultSucceeded
		intent.result.Message = fmt.Sprintf("%s constructed", intent.building)

	case game.ActionAttack:
		if action.Attack == nil {
			return intent.invalid("Attack payload was missing")
		}
		intent.from, intent.to = action.Attack.From, action.Attack.To
		source, target := m.GetCell(intent.from), m.GetCell(intent.to)
		if source == nil || target == nil ||
			!source.HasUnits() ||
			source.Units[0].Type == game.UnitScout ||
			source.Units[0].Owner != factionOwner {
			return intent.invalid("A friendly non-Scout unit is required to attack")
		}
		if !game.HexAdjacent(intent.from, intent.to) {
			return intent.invalid("Attack target must be adjacent")
		}
		if !hasEnemies(target, factionOwner) {
			return intent.invalid("Target has no enemy units or buildings")
		}
		intent.targetsTile = true
		intent.valid = true
		intent.result.Status = game.ActionResultSucceeded
		intent.result.Message = "Attack ordered"

	default:
		return intent.invalid("Unknown action")
	}
	return intent
}

type moveCandidate struct {
	isAttack bool
	from     game.Hex
	dest     game.Hex
	unitType game.UnitType
	budget   int
}

func (gi *GameInstance) planAutomaticMovement(factionIdx int) *roundIntent {
	factionOwner := int8(factionIdx)

	// Phase 3: build combined candidate list from attack + movement orders
	allAttacks := gi.attackOrders[factionIdx]
	var remainingAttacks []game.AttackOrder
	var attackCands, moveCands []moveCandidate

	for _, order := range allAttacks {
		source := gi.game.Map.GetCell(order.From)
		if source == nil || !source.HasUnits() || source.Units[0].Owner != factionOwner {
			continue
		}
		if game.HexAdjacent(order.From, order.TargetTile) {
			remainingAttacks = append(remainingAttacks, order)
			continue
		}
		remainingAttacks = append(remainingAttacks, order)
		attackCands = append(attackCands, moveCandidate{
			isAttack: true,
			from:     order.From,
			dest:     order.TargetTile,
			unitType: source.Units[0].Type,
			budget:   game.UnitMovementBudget(source.Units[0].Type),
		})
	}
	gi.attackOrders[factionIdx] = remainingAttacks

	movementOrders := gi.movementOrders[factionIdx]
	if priority, ok := gi.movementPriorities[factionIdx]; ok {
		for i, order := range movementOrders {
			if order.Current != priority {
				continue
			}
			prioritized := order
			copy(movementOrders[1:i+1], movementOrders[0:i])
			movementOrders[0] = prioritized
			break
		}
	}
	var remainingMoves []game.MovementOrder
	for _, order := range movementOrders {
		source := gi.game.Map.GetCell(order.Current)
		if source == nil || !source.HasUnits() || source.Units[0].Owner != factionOwner ||
			order.Current == order.Destination {
			continue
		}
		remainingMoves = append(remainingMoves, order)
		moveCands = append(moveCands, moveCandidate{
			isAttack: false,
			from:     order.Current,
			dest:     order.Destination,
			unitType: source.Units[0].Type,
			budget:   game.UnitMovementBudget(source.Units[0].Type),
		})
	}

	// Interleave round-robin: alternate which type goes first each round
	var candidates []moveCandidate
	if gi.game.Round%2 == 0 {
		candidates = append(moveCands, attackCands...)
	} else {
		candidates = append(attackCands, moveCands...)
	}

	for _, cand := range candidates {
		if cand.isAttack {
			path, _, ok := gi.game.Map.FindAdjacentApproachPath(
				factionOwner, cand.from, cand.dest,
			)
			if !ok {
				continue
			}
			traversed := gi.game.Map.AdvanceUnitPath(path, cand.budget)
			if len(traversed) < 2 {
				continue
			}
			gi.attackOrders[factionIdx] = removeAttackOrder(
				gi.attackOrders[factionIdx], cand.from,
			)
			gi.attackOrders[factionIdx] = append(
				gi.attackOrders[factionIdx], game.AttackOrder{
					From:       traversed[len(traversed)-1],
					TargetTile: cand.dest,
				},
			)
			intent := gi.newIntent(factionIdx, game.ActionMove, true)
			intent.from = cand.from
			intent.destination = cand.dest
			intent.to = traversed[len(traversed)-1]
			intent.unit = cand.unitType
			intent.path = traversed
			intent.targetsTile = true
			intent.valid = true
			intent.result.Status = game.ActionResultSucceeded
			intent.result.Message = "Advancing toward target"
			return intent
		}
		path, ok := gi.game.Map.FindUnitPath(factionOwner, cand.from, cand.dest)
		if !ok {
			continue
		}
		traversed := gi.game.Map.AdvanceUnitPath(path, cand.budget)
		if len(traversed) < 2 {
			continue
		}
		// Remove old order — applyIntent will append the new one
		var updated []game.MovementOrder
		for _, o := range gi.movementOrders[factionIdx] {
			if o.Current != cand.from {
				updated = append(updated, o)
			}
		}
		gi.movementOrders[factionIdx] = updated
		intent := gi.newIntent(factionIdx, game.ActionMove, true)
		intent.from = cand.from
		intent.destination = cand.dest
		intent.to = traversed[len(traversed)-1]
		intent.unit = cand.unitType
		intent.path = traversed
		intent.targetsTile = true
		intent.valid = true
		intent.result.Status = game.ActionResultSucceeded
		intent.result.Message = "Queued unit advanced"
		return intent
	}

	// No movement possible — save remaining orders
	gi.movementOrders[factionIdx] = remainingMoves
	if len(remainingAttacks) == 0 && len(remainingMoves) == 0 {
		return nil
	}
	intent := gi.newIntent(factionIdx, game.ActionMove, true)
	intent.result.Status = game.ActionResultBlocked
	intent.result.Message = "All queued routes are blocked"
	return intent
}

func (gi *GameInstance) applyIntent(intent *roundIntent) {
	faction := &gi.game.Factions[intent.faction]
	owner := int8(intent.faction)
	m := &gi.game.Map

	switch intent.action {
	case game.ActionMove:
		source := m.GetCell(intent.from)
		target := m.GetCell(intent.to)
		source.Units = nil
		stats := game.GetUnitStats(intent.unit)
		target.Units = []game.UnitData{{Type: intent.unit, Owner: owner, HP: stats.MaxHP}}

		if intent.destination != intent.to {
			gi.movementOrders[intent.faction] = append(
				gi.movementOrders[intent.faction],
				game.MovementOrder{Current: intent.to, Destination: intent.destination},
			)
		}
		gi.movementEvents = append(gi.movementEvents, game.MovementEvent{
			Unit:  intent.unit,
			Owner: owner,
			Path:  append([]game.Hex(nil), intent.path...),
		})

	case game.ActionRecruit:
		target := m.GetCell(intent.to)
		faction.Coins -= intent.cost
		for res, amt := range game.UnitResourceCost(intent.unit) {
			faction.Resources[res] -= amt
		}
		stats := game.GetUnitStats(intent.unit)
		target.Units = []game.UnitData{{Type: intent.unit, Owner: owner, HP: stats.MaxHP}}

	case game.ActionBuild:
		target := m.GetCell(intent.to)
		faction.Coins -= intent.cost
		target.Owner = owner
		maxHP := game.BuildingMaxHP(intent.building)
		target.Building = &game.BuildingData{Type: intent.building, HP: maxHP}

	case game.ActionAttack:
		gi.resolveAttack(intent.faction, intent.from, intent.to)
	}
}

func (gi *GameInstance) resolveAttack(factionIdx int, from, to game.Hex) bool {
	source := gi.game.Map.GetCell(from)
	target := gi.game.Map.GetCell(to)
	if source == nil || target == nil {
		return false
	}
	if !source.HasUnits() {
		return false
	}

	gi.attackEvents = append(gi.attackEvents, game.AttackEvent{
		Unit:   source.Units[0].Type,
		Owner:  int8(factionIdx),
		From:   from,
		Target: to,
	})

	unitOwner := int8(factionIdx)
	stats := game.GetUnitStats(source.Units[0].Type)
	attackPower := stats.Attack

	if target.HasUnits() {
		var enemyIdx int
		enemyFound := false
		for i, u := range target.Units {
			if u.Owner != unitOwner {
				enemyIdx = i
				enemyFound = true
				break
			}
		}
		if !enemyFound {
			goto tryBuilding
		}
		target.Units[enemyIdx].HP -= int8(attackPower)
		if target.Units[enemyIdx].HP <= 0 {
			copy(target.Units[enemyIdx:], target.Units[enemyIdx+1:])
			target.Units = target.Units[:len(target.Units)-1]
		}
		return true
	}

tryBuilding:
	if target.HasBuilding() && target.Owner >= 0 && target.Owner != unitOwner {
		target.Building.HP -= int8(attackPower)
		if target.Building.HP <= 0 {
			target.Building = nil
			target.Owner = -1
		}
		return true
	}

	return false
}

func (gi *GameInstance) newIntent(factionIdx int, action game.ActionType, automatic bool) *roundIntent {
	return &roundIntent{
		faction:   factionIdx,
		action:    action,
		automatic: automatic,
		result: game.ActionResult{
			Round:     gi.game.Round,
			Type:      action,
			Automatic: automatic,
		},
	}
}

func (intent *roundIntent) invalid(message string) *roundIntent {
	intent.result.Status = game.ActionResultInvalid
	intent.result.Message = message
	return intent
}

func (intent *roundIntent) insufficientCoins() *roundIntent {
	intent.result.Status = game.ActionResultInsufficientCoins
	intent.result.Message = "Not enough coins"
	return intent
}

func (intent *roundIntent) insufficientResources(res game.ResourceType) *roundIntent {
	intent.result.Status = game.ActionResultInvalid
	intent.result.Message = fmt.Sprintf("Not enough %s", res.String())
	return intent
}

func (gi *GameInstance) recordResult(intent *roundIntent) {
	intent.result.From = intent.from
	if intent.action == game.ActionMove {
		intent.result.To = intent.destination
	} else {
		intent.result.To = intent.to
	}
	result := intent.result
	gi.actionResults[intent.faction] = &result
}

func removeAttackOrder(orders []game.AttackOrder, from game.Hex) []game.AttackOrder {
	filtered := orders[:0]
	for _, order := range orders {
		if order.From != from {
			filtered = append(filtered, order)
		}
	}
	return filtered
}

func removeMovementOrder(orders []game.MovementOrder, current game.Hex) []game.MovementOrder {
	filtered := orders[:0]
	for _, order := range orders {
		if order.Current != current {
			filtered = append(filtered, order)
		}
	}
	return filtered
}
