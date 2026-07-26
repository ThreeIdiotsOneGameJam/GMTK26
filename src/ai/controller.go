package ai

import (
	"fmt"
	"math"

	"github.com/threeidiotsonegamejam/gmtk26/src/game"
)

type Controller struct {
	mapSeed     int64
	owner       int8
	config      Config
	personality Personality

	decisionCount int32
	currentGoal   Goal
	goalSince     int32
	hasGoal       bool

	target      targetRef
	targetSince int32
	hasTarget   bool

	opponents     [4]opponentMemory
	previousOwned map[game.Hex]bool
	cooldowns     map[game.Hex]int32
	recentActions []string

	// combatAdvanceDebt reserves future action-free rounds after development
	// so building and recruitment cannot indefinitely freeze an active army.
	combatAdvanceDebt int
}

func NewController(mapSeed int64, owner int8, config Config) *Controller {
	config = normalizeConfig(config)
	return &Controller{
		mapSeed:       mapSeed,
		owner:         owner,
		config:        config,
		personality:   seededPersonality(mapSeed, owner),
		previousOwned: make(map[game.Hex]bool),
		cooldowns:     make(map[game.Hex]int32),
	}
}

func normalizeConfig(config Config) Config {
	defaults := StandardConfig()
	if config.MaxCandidates <= 0 {
		config.MaxCandidates = defaults.MaxCandidates
	}
	if config.MaxDetailedCandidates <= 0 {
		config.MaxDetailedCandidates = defaults.MaxDetailedCandidates
	}
	if config.MaxPathQueries <= 0 {
		config.MaxPathQueries = defaults.MaxPathQueries
	}
	if config.ForecastRounds <= 0 {
		config.ForecastRounds = defaults.ForecastRounds
	}
	if config.GoalCommitRounds <= 0 {
		config.GoalCommitRounds = defaults.GoalCommitRounds
	}
	if config.TargetCommitRounds <= 0 {
		config.TargetCommitRounds = defaults.TargetCommitRounds
	}
	if config.EmergencyThreat <= 0 {
		config.EmergencyThreat = defaults.EmergencyThreat
	}
	if config.GoalSwitchMargin <= 0 {
		config.GoalSwitchMargin = defaults.GoalSwitchMargin
	}
	if config.TargetSwitchMultiplier <= 1 {
		config.TargetSwitchMultiplier = defaults.TargetSwitchMultiplier
	}
	if config.ChoiceBand <= 0 {
		config.ChoiceBand = defaults.ChoiceBand
	}
	if config.ChoiceTemperature <= 0 {
		config.ChoiceTemperature = defaults.ChoiceTemperature
	}
	return config
}

func (controller *Controller) Plan(world *WorldSnapshot, own OwnState) Plan {
	if world == nil || own.Owner != controller.owner ||
		!validOwner(own.Owner) || !world.Factions[own.Owner].Alive {
		return Plan{}
	}
	controller.observe(world, own)
	analysis := analyzeWorld(
		world,
		own,
		controller.opponents,
		controller.config.ForecastRounds,
	)
	controller.selectTarget(world, analysis)

	scores := goalScores(analysis, controller.personality)
	goal, goalUtility := controller.selectGoal(scores, analysis)
	oracle := newPathOracle(world, own.Owner, controller.config.MaxPathQueries)
	orders, routeValue := planRoutes(
		world,
		own,
		analysis,
		goal,
		controller.target,
		controller.hasTarget,
		oracle,
		controller.config,
	)
	analysis.signals.routeValue = math.Max(analysis.signals.routeValue, routeValue)

	candidates := generateCandidates(
		world,
		own,
		analysis,
		goal,
		controller.personality,
		controller.recentActions,
		controller.config.MaxCandidates,
	)
	selectionCandidates := candidates
	hasCombatRoute := hasNonAdjacentCombatRoute(own, orders)
	if attacks := immediateAttackCandidates(candidates); len(attacks) > 0 {
		selectionCandidates = attacks
	} else if controller.combatAdvanceDebt > 0 && hasCombatRoute {
		if advances := advanceCandidates(candidates); len(advances) > 0 {
			selectionCandidates = advances
		}
	}
	choice, alternatives := chooseCandidate(
		selectionCandidates,
		uint64(controller.mapSeed)^
			uint64(uint8(controller.owner)+1)<<48^
			uint64(world.Round)<<16^
			uint64(controller.decisionCount),
		controller.config.ChoiceBand,
		controller.config.ChoiceTemperature,
	)
	controller.updateCombatCadence(
		choice,
		hasCombatRoute,
		len(analysis.ownUnits)-analysis.unitCounts[game.UnitScout],
	)
	orders = filterOrdersForManual(orders, choice.manual)

	if choice.key != "" {
		controller.recentActions = append(controller.recentActions, choice.key)
		if len(controller.recentActions) > 3 {
			controller.recentActions = controller.recentActions[len(controller.recentActions)-3:]
		}
	}
	controller.rememberOwned(analysis)
	controller.decisionCount++

	trace := DecisionTrace{
		Round:              world.Round,
		Faction:            own.Owner,
		Goal:               goal,
		GoalUtility:        goalUtility,
		Choice:             choice.description,
		ChoiceUtility:      choice.utility,
		Alternatives:       alternatives,
		Candidates:         len(candidates),
		DetailedCandidates: min(len(candidates), controller.config.MaxDetailedCandidates),
		PathQueries:        oracle.queries,
		Personality:        controller.personality,
	}
	if controller.hasTarget {
		trace.Target = controller.target.Position
		trace.TargetOwner = controller.target.Owner
		trace.HasTarget = true
	}
	return Plan{
		Manual: choice.manual,
		Orders: orders,
		Trace:  trace,
	}
}

func immediateAttackCandidates(candidates []candidate) []candidate {
	attacks := make([]candidate, 0)
	for _, item := range candidates {
		if item.manual != nil && item.manual.Type == game.ActionAttack {
			attacks = append(attacks, item)
		}
	}
	return attacks
}

func advanceCandidates(candidates []candidate) []candidate {
	for _, item := range candidates {
		if item.key == "advance" && item.manual == nil {
			return []candidate{item}
		}
	}
	return nil
}

func hasNonAdjacentCombatRoute(own OwnState, commands []OrderCommand) bool {
	for _, order := range own.AttackOrders {
		if !game.HexAdjacent(order.From, order.TargetTile) {
			return true
		}
	}
	for _, command := range commands {
		if command.Kind == OrderAttack &&
			!game.HexAdjacent(command.From, command.To) {
			return true
		}
	}
	return false
}

func (controller *Controller) updateCombatCadence(
	choice candidate,
	hasCombatRoute bool,
	combatUnits int,
) {
	if !hasCombatRoute {
		controller.combatAdvanceDebt = 0
		return
	}
	if choice.manual == nil {
		if controller.combatAdvanceDebt > 0 {
			controller.combatAdvanceDebt--
		}
		return
	}
	switch choice.manual.Type {
	case game.ActionBuild, game.ActionRecruit:
		required := min(3, max(1, (combatUnits+2)/3))
		controller.combatAdvanceDebt = max(
			controller.combatAdvanceDebt,
			required,
		)
	case game.ActionAttack:
		if controller.combatAdvanceDebt > 0 {
			controller.combatAdvanceDebt--
		}
	}
}

func filterOrdersForManual(orders []OrderCommand, manual *ManualAction) []OrderCommand {
	if manual == nil {
		return orders
	}
	filtered := orders[:0]
	for _, order := range orders {
		conflict := false
		switch manual.Type {
		case game.ActionBuild:
			conflict = manual.Build != nil &&
				(order.From == manual.Build.From || order.To == manual.Build.To)
		case game.ActionRecruit:
			conflict = manual.Recruit != nil && order.To == manual.Recruit.To
		case game.ActionAttack:
			conflict = manual.Attack != nil && order.From == manual.Attack.From
		}
		if !conflict {
			filtered = append(filtered, order)
		}
	}
	return filtered
}

func (controller *Controller) selectGoal(
	scores [8]float64,
	analysis worldAnalysis,
) (Goal, float64) {
	best, bestScore := highestGoal(scores)
	if !controller.hasGoal {
		controller.currentGoal = best
		controller.goalSince = controller.decisionCount
		controller.hasGoal = true
		return best, bestScore
	}
	if analysis.signals.townhallThreat >= controller.config.EmergencyThreat {
		controller.currentGoal = GoalDefend
		controller.goalSince = controller.decisionCount
		return GoalDefend, scores[GoalDefend]
	}

	currentScore := clamp01(scores[controller.currentGoal] + 0.08)
	committed := controller.decisionCount-controller.goalSince <
		controller.config.GoalCommitRounds
	if !committed && best != controller.currentGoal &&
		bestScore > currentScore+controller.config.GoalSwitchMargin {
		controller.currentGoal = best
		controller.goalSince = controller.decisionCount
		return best, bestScore
	}
	return controller.currentGoal, currentScore
}

func (controller *Controller) selectTarget(
	world *WorldSnapshot,
	analysis worldAnalysis,
) {
	if controller.hasTarget {
		valid := false
		currentScore := 0.0
		for _, target := range analysis.targets {
			if target.Position == controller.target.Position &&
				target.Owner == controller.target.Owner {
				controller.target = target
				currentScore = target.Score
				valid = true
				break
			}
		}
		if !valid {
			controller.hasTarget = false
		} else if controller.decisionCount-controller.targetSince <
			controller.config.TargetCommitRounds {
			return
		} else if len(analysis.targets) == 0 ||
			analysis.targets[0].Score <= currentScore*controller.config.TargetSwitchMultiplier {
			return
		}
	}
	for _, target := range analysis.targets {
		if until := controller.cooldowns[target.Position]; until >= world.Round {
			continue
		}
		controller.target = target
		controller.targetSince = controller.decisionCount
		controller.hasTarget = true
		return
	}
	controller.hasTarget = false
}

func (controller *Controller) observe(world *WorldSnapshot, own OwnState) {
	for i := range controller.opponents {
		controller.opponents[i].Retaliation *= 0.8
		if i == int(controller.owner) || !world.Factions[i].Alive {
			continue
		}
		delta := float64(world.Factions[i].Points - controller.opponents[i].LastPoints)
		controller.opponents[i].ScoreVelocity =
			0.35*delta + 0.65*controller.opponents[i].ScoreVelocity
		controller.opponents[i].LastPoints = world.Factions[i].Points
	}
	for _, attack := range world.Attacks {
		if attack.Owner == controller.owner || !validOwner(attack.Owner) {
			continue
		}
		if controller.previousOwned[attack.Target] {
			controller.opponents[attack.Owner].Retaliation = clamp01(
				controller.opponents[attack.Owner].Retaliation + 0.3,
			)
		}
	}
	if own.LastResult != nil &&
		(own.LastResult.Status == game.ActionResultContested ||
			own.LastResult.Status == game.ActionResultBlocked) {
		controller.cooldowns[own.LastResult.To] = world.Round + 2
	}
	for position, until := range controller.cooldowns {
		if until < world.Round {
			delete(controller.cooldowns, position)
		}
	}
}

func (controller *Controller) rememberOwned(analysis worldAnalysis) {
	controller.previousOwned = make(map[game.Hex]bool)
	for _, building := range analysis.ownBuildings {
		controller.previousOwned[building.Position] = true
	}
	for _, unit := range analysis.ownUnits {
		controller.previousOwned[unit.Position] = true
	}
}

func (trace DecisionTrace) String() string {
	target := "-"
	if trace.HasTarget {
		target = fmt.Sprintf(
			"f%d@%d,%d",
			trace.TargetOwner,
			trace.Target.X,
			trace.Target.Y,
		)
	}
	return fmt.Sprintf(
		"round=%d faction=%d goal=%s(%.2f) choice=%q(%.2f) target=%s candidates=%d paths=%d",
		trace.Round,
		trace.Faction,
		trace.Goal,
		trace.GoalUtility,
		trace.Choice,
		trace.ChoiceUtility,
		target,
		trace.Candidates,
		trace.PathQueries,
	)
}
