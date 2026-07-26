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
	troop       game.TroopType
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
					gi.movementOrders[intent.faction],
					intent.from,
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
		intent.from, intent.to, intent.troop = action.Recruit.From, action.Recruit.To, action.Recruit.Troop
		source, target := m.GetCell(intent.from), m.GetCell(intent.to)
		if source == nil || target == nil || source.Owner != factionOwner {
			return intent.invalid("Recruitment source is not friendly")
		}
		if !game.HexAdjacent(intent.from, intent.to) {
			return intent.invalid("Recruitment target must be adjacent")
		}
		if game.TerrainMovementCost(target.Tile) <= 0 ||
			(target.Owner != -1 && target.Owner != factionOwner) ||
			target.Troop != game.TroopUnknown {
			return intent.invalid("Recruitment target is not available")
		}
		validTroop := intent.troop >= game.TroopPeasant && intent.troop <= game.TroopScout
		if !validTroop ||
			source.Building == game.BuildingTownhall && intent.troop != game.TroopScout ||
			source.Building != game.BuildingTownhall && source.Building != game.BuildingBarracks {
			return intent.invalid("Building cannot recruit that troop")
		}
		intent.cost = game.TroopCost(intent.troop)
		if faction.Coins < intent.cost {
			return intent.insufficientCoins()
		}
		intent.targetsTile = true
		intent.valid = true
		intent.result.Status = game.ActionResultSucceeded
		intent.result.Message = fmt.Sprintf("%s recruited", intent.troop)

	case game.ActionBuild:
		if action.Build == nil {
			return intent.invalid("Build payload was missing")
		}
		intent.from, intent.to, intent.building = action.Build.From, action.Build.To, action.Build.Building
		source, target := m.GetCell(intent.from), m.GetCell(intent.to)
		if source == nil || target == nil ||
			source.Troop != game.TroopScout ||
			source.TroopOwner != factionOwner {
			return intent.invalid("A friendly Scout is required to build")
		}
		if intent.from != intent.to && !game.HexAdjacent(intent.from, intent.to) {
			return intent.invalid("Scout may only build on its tile or an adjacent tile")
		}
		if target.Owner != -1 && target.Owner != factionOwner {
			return intent.invalid("Cannot build in enemy territory")
		}
		if target.Troop != game.TroopUnknown && target.TroopOwner != factionOwner {
			return intent.invalid("Cannot build beneath an enemy troop")
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
			source.Troop == game.TroopUnknown ||
			source.Troop == game.TroopScout ||
			source.TroopOwner != factionOwner {
			return intent.invalid("A friendly non-Scout troop is required to attack")
		}
		if !game.HexAdjacent(intent.from, intent.to) {
			return intent.invalid("Attack target must be adjacent")
		}
		if target.Owner == -1 || target.Owner == factionOwner ||
			target.Building == game.BuildingUnknown ||
			target.Building == game.BuildingTownhall {
			return intent.invalid("Building is not a valid attack target")
		}
		intent.valid = true
		intent.result.Status = game.ActionResultSucceeded
		intent.result.Message = fmt.Sprintf("%s demolished", target.Building)

	default:
		return intent.invalid("Unknown action")
	}
	return intent
}

func (gi *GameInstance) planAutomaticMovement(factionIdx int) *roundIntent {
	orders := gi.movementOrders[factionIdx]
	if priority, ok := gi.movementPriorities[factionIdx]; ok {
		for i, order := range orders {
			if order.Current != priority {
				continue
			}
			prioritized := order
			copy(orders[1:i+1], orders[0:i])
			orders[0] = prioritized
			break
		}
	}
	checked := len(orders)
	factionOwner := int8(factionIdx)

	for range checked {
		order := orders[0]
		orders = orders[1:]
		source := gi.game.Map.GetCell(order.Current)
		if source == nil || source.Troop == game.TroopUnknown || source.TroopOwner != factionOwner ||
			order.Current == order.Destination {
			continue
		}
		path, ok := gi.game.Map.FindTroopPath(factionOwner, order.Current, order.Destination)
		if !ok {
			orders = append(orders, order)
			continue
		}
		traversed := gi.game.Map.AdvanceTroopPath(path, game.TroopMovementBudget(source.Troop))
		if len(traversed) < 2 {
			orders = append(orders, order)
			continue
		}

		gi.movementOrders[factionIdx] = orders
		intent := gi.newIntent(factionIdx, game.ActionMove, true)
		intent.from = order.Current
		intent.destination = order.Destination
		intent.to = traversed[len(traversed)-1]
		intent.troop = source.Troop
		intent.path = traversed
		intent.targetsTile = true
		intent.valid = true
		intent.result.Status = game.ActionResultSucceeded
		intent.result.Message = "Queued troop advanced"
		return intent
	}

	gi.movementOrders[factionIdx] = orders
	if len(orders) == 0 {
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
		source.Troop = game.TroopUnknown
		source.TroopOwner = 0
		target.Troop = intent.troop
		target.TroopOwner = owner

		if intent.destination != intent.to {
			gi.movementOrders[intent.faction] = append(
				gi.movementOrders[intent.faction],
				game.MovementOrder{Current: intent.to, Destination: intent.destination},
			)
		}
		gi.movementEvents = append(gi.movementEvents, game.MovementEvent{
			Troop: intent.troop,
			Owner: owner,
			Path:  append([]game.Hex(nil), intent.path...),
		})

	case game.ActionRecruit:
		target := m.GetCell(intent.to)
		faction.Coins -= intent.cost
		target.Troop = intent.troop
		target.TroopOwner = owner

	case game.ActionBuild:
		target := m.GetCell(intent.to)
		faction.Coins -= intent.cost
		target.Owner = owner
		target.Building = intent.building

	case game.ActionAttack:
		target := m.GetCell(intent.to)
		target.Building = game.BuildingUnknown
		target.Owner = -1
	}
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

func removeMovementOrder(orders []game.MovementOrder, current game.Hex) []game.MovementOrder {
	filtered := orders[:0]
	for _, order := range orders {
		if order.Current != current {
			filtered = append(filtered, order)
		}
	}
	return filtered
}
