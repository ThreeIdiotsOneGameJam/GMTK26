package server

import (
	"log"
	"os"

	gameai "github.com/threeidiotsonegamejam/gmtk26/src/ai"
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
)

var aiDebug = os.Getenv("AI_DEBUG") == "1"

func (gi *GameInstance) initializeAIControllers() {
	gi.mu.Lock()
	defer gi.mu.Unlock()
	gi.ensureAIControllersLocked()
}

func (gi *GameInstance) ensureAIControllersLocked() {
	for i := range gi.game.Factions {
		faction := &gi.game.Factions[i]
		if !faction.AI || !faction.Alive {
			continue
		}
		if _, exists := gi.aiControllers[i]; exists {
			continue
		}
		gi.aiControllers[i] = gameai.NewController(
			gi.game.Map.Seed,
			int8(i),
			gameai.StandardConfig(),
		)
	}
}

func (gi *GameInstance) activateAITakeoversLocked() {
	for factionIdx := range gi.pendingAITakeovers {
		if factionIdx < 0 || factionIdx >= len(gi.game.Factions) {
			delete(gi.pendingAITakeovers, factionIdx)
			continue
		}
		faction := &gi.game.Factions[factionIdx]
		faction.AI = true
		delete(gi.actions, factionIdx)
		delete(gi.movementPriorities, factionIdx)
		if faction.Alive {
			gi.aiControllers[factionIdx] = gameai.NewController(
				gi.game.Map.Seed,
				int8(factionIdx),
				gameai.StandardConfig(),
			)
		}
		delete(gi.pendingAITakeovers, factionIdx)
	}
}

func (gi *GameInstance) resolveRoundLocked() int {
	gi.activateAITakeoversLocked()
	gi.ensureAIControllersLocked()
	gi.planAIActionsLocked()
	gi.processAutoActions()
	gi.processClientActions()
	aliveCount := gi.checkAlive()
	if controlScoreDue(gi.game.Round) {
		gi.awardControlScore()
	}
	gi.game.Round++
	gi.actions = make(map[int]*submittedAction)
	return aliveCount
}

func (gi *GameInstance) planAIActionsLocked() {
	snapshot := gameai.NewWorldSnapshot(
		gi.game,
		gi.movementEvents,
		gi.attackEvents,
	)
	plans := make(map[int]gameai.Plan)
	for factionIdx := range gi.game.Factions {
		faction := &gi.game.Factions[factionIdx]
		planner, exists := gi.aiControllers[factionIdx]
		if !exists || !faction.AI || !faction.Alive {
			continue
		}
		own := gameai.OwnState{
			Owner: int8(factionIdx),
			MovementOrders: append(
				[]game.MovementOrder(nil),
				gi.movementOrders[factionIdx]...,
			),
			AttackOrders: append(
				[]game.AttackOrder(nil),
				gi.attackOrders[factionIdx]...,
			),
		}
		if result := gi.actionResults[factionIdx]; result != nil {
			copied := *result
			own.LastResult = &copied
		}
		plan, ok := safelyPlanAI(planner, &snapshot, own)
		if ok {
			plans[factionIdx] = plan
		}
	}

	// Apply only after every planner has inspected the same resolved snapshot.
	for factionIdx := range gi.game.Factions {
		plan, exists := plans[factionIdx]
		if !exists {
			continue
		}
		gi.applyAIPlanLocked(factionIdx, plan)
		gi.aiTraces[factionIdx] = plan.Trace
		if aiDebug {
			log.Printf("game %d AI %s", gi.ID, plan.Trace.String())
		}
	}
}

func safelyPlanAI(
	planner gameai.Planner,
	snapshot *gameai.WorldSnapshot,
	own gameai.OwnState,
) (plan gameai.Plan, ok bool) {
	defer func() {
		if recovered := recover(); recovered != nil {
			log.Printf("AI faction %d planning panic: %v", own.Owner, recovered)
			ok = false
		}
	}()
	return planner.Plan(snapshot, own), true
}

func (gi *GameInstance) applyAIPlanLocked(
	factionIdx int,
	plan gameai.Plan,
) {
	var primary *game.Hex
	for _, command := range plan.Orders {
		var err error
		switch command.Kind {
		case gameai.OrderMove:
			err = gi.setMovementOrderLocked(
				factionIdx,
				game.MoveActionPayload{From: command.From, To: command.To},
			)
			if err == nil && primary == nil {
				from := command.From
				primary = &from
			}
		case gameai.OrderAttack:
			err = gi.setAttackOrderLocked(
				factionIdx,
				game.AttackActionPayload{From: command.From, To: command.To},
			)
		default:
			err = errInvalidAIOrder
		}
		if err != nil && aiDebug {
			log.Printf(
				"game %d: rejected AI faction %d order %+v: %v",
				gi.ID,
				factionIdx,
				command,
				err,
			)
		}
	}
	if primary != nil {
		gi.movementPriorities[factionIdx] = *primary
	}

	delete(gi.actions, factionIdx)
	if plan.Manual == nil {
		return
	}
	action := plan.Manual
	switch action.Type {
	case game.ActionPass:
		gi.actions[factionIdx] = &submittedAction{Type: game.ActionPass}
	case game.ActionBuild:
		if action.Build != nil {
			payload := *action.Build
			gi.actions[factionIdx] = &submittedAction{
				Type:  game.ActionBuild,
				Build: &payload,
			}
		}
	case game.ActionRecruit:
		if action.Recruit != nil {
			payload := *action.Recruit
			gi.actions[factionIdx] = &submittedAction{
				Type:    game.ActionRecruit,
				Recruit: &payload,
			}
		}
	case game.ActionAttack:
		if action.Attack != nil && game.HexAdjacent(action.Attack.From, action.Attack.To) {
			payload := *action.Attack
			gi.actions[factionIdx] = &submittedAction{
				Type:   game.ActionAttack,
				Attack: &payload,
			}
		}
	}
}

type invalidAIOrderError struct{}

func (invalidAIOrderError) Error() string { return "invalid AI order kind" }

var errInvalidAIOrder error = invalidAIOrderError{}
