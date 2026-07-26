package ai

import (
	"fmt"
	"math"

	"github.com/threeidiotsonegamejam/gmtk26/src/game"
)

func generateCandidates(
	world *WorldSnapshot,
	own OwnState,
	analysis worldAnalysis,
	goal Goal,
	personality Personality,
	recent []string,
	limit int,
) []candidate {
	if limit < 2 {
		limit = 2
	}
	candidates := make([]candidate, 0, min(limit, 64))
	add := func(item candidate) bool {
		if len(candidates) >= limit {
			return false
		}
		item.utility = clamp01(item.utility)
		candidates = append(candidates, item)
		return true
	}

	routeValue := analysis.signals.routeValue
	add(candidate{
		key:         "advance",
		description: "Advance best route",
		utility: candidateUtility(
			routeGoalFit(goal),
			0.10*routeValue,
			routeValue,
			math.Max(analysis.signals.townhallThreat, analysis.signals.scorePressure),
			0.85,
			0.75,
			novelty("advance", recent),
			0,
			0,
		),
	})
	pass := &ManualAction{Type: game.ActionPass}
	add(candidate{
		key:         "pass",
		description: "Hold position",
		manual:      pass,
		utility: candidateUtility(
			0.05,
			0,
			0.05,
			0,
			1,
			0.80,
			novelty("pass", recent),
			0,
			0.10*routeValue,
		),
	})

	existingAttacks := indexAttackOrders(own.AttackOrders)
	for _, unit := range analysis.ownUnits {
		if len(candidates) >= limit {
			break
		}
		if unit.Data.Type == game.UnitScout {
			generateBuildCandidates(
				world,
				analysis,
				goal,
				personality,
				recent,
				unit,
				routeValue,
				add,
			)
			continue
		}
		generateAttackCandidates(
			world,
			analysis,
			goal,
			personality,
			recent,
			unit,
			existingAttacks,
			routeValue,
			add,
		)
	}

	for _, building := range analysis.ownBuildings {
		if len(candidates) >= limit {
			break
		}
		if building.Data.Type != game.BuildingTownhall &&
			building.Data.Type != game.BuildingBarracks {
			continue
		}
		generateRecruitCandidates(
			world,
			analysis,
			goal,
			personality,
			recent,
			building,
			routeValue,
			add,
		)
	}
	return candidates
}

func generateBuildCandidates(
	world *WorldSnapshot,
	analysis worldAnalysis,
	goal Goal,
	personality Personality,
	recent []string,
	scout unitRef,
	routeValue float64,
	add func(candidate) bool,
) {
	positions := append([]game.Hex{scout.Position}, neighborHexes(scout.Position)...)
	buildings := [...]game.BuildingType{
		game.BuildingFarm,
		game.BuildingForester,
		game.BuildingMine,
		game.BuildingBarracks,
		game.BuildingBank,
	}
	for _, position := range positions {
		cell := world.Map.GetCell(position)
		if cell == nil {
			continue
		}
		for _, building := range buildings {
			payload := game.BuildActionPayload{
				From:     scout.Position,
				To:       position,
				Building: building,
			}
			validation := game.ValidateBuildAction(
				&world.Map,
				analysis.owner,
				payload,
				analysis.funds,
			)
			if !validation.Valid {
				continue
			}
			key := fmt.Sprintf(
				"build:%d:%d:%d",
				building,
				position.X,
				position.Y,
			)
			siteValue := analysis.siteScore(position, building, cell.Tile)
			controlNow := 0.0
			if world.Round%game.ScoreIntervalRounds == 0 {
				controlNow = float64(game.BuildingControlScore(building, cell.Tile)) / 5
			}
			urgency := buildingUrgency(analysis, building, cell.Tile)
			goalFit := buildGoalFit(goal, building)
			efficiency := costEfficiency(
				validation.CoinCost,
				validation.ResourceCost,
				analysis,
			)
			safety := localSafety(analysis, position)
			reserve := reservePenalty(validation.ResourceCost, analysis)
			utility := candidateUtility(
				goalFit,
				controlNow,
				siteValue,
				urgency,
				efficiency,
				safety,
				novelty(key, recent),
				0.12*reserve,
				0.10*routeValue,
			)
			utility *= 0.5*personality.Economy + 0.5*personality.Expansion
			if !add(candidate{
				key:         key,
				description: fmt.Sprintf("Build %s at %d,%d", building, position.X, position.Y),
				manual: &ManualAction{
					Type:  game.ActionBuild,
					Build: &payload,
				},
				utility: utility,
			}) {
				return
			}
		}
	}
}

func generateRecruitCandidates(
	world *WorldSnapshot,
	analysis worldAnalysis,
	goal Goal,
	personality Personality,
	recent []string,
	building buildingRef,
	routeValue float64,
	add func(candidate) bool,
) {
	units := [...]game.UnitType{
		game.UnitScout,
		game.UnitPeasant,
		game.UnitArcher,
		game.UnitKnight,
	}
	for _, position := range neighborHexes(building.Position) {
		for _, unit := range units {
			payload := game.RecruitActionPayload{
				From: building.Position,
				To:   position,
				Unit: unit,
			}
			validation := game.ValidateRecruitAction(
				&world.Map,
				analysis.owner,
				payload,
				analysis.funds,
			)
			if !validation.Valid {
				continue
			}
			key := fmt.Sprintf(
				"recruit:%d:%d:%d",
				unit,
				position.X,
				position.Y,
			)
			stats := game.GetUnitStats(unit)
			power := float64(stats.MaxHP) +
				2*float64(stats.Attack) +
				0.5*float64(game.UnitMovementBudget(unit))
			horizon := clamp01(power / 16)
			urgency := math.Max(
				analysis.signals.militaryGap,
				analysis.signals.townhallThreat,
			)
			if unit == game.UnitScout {
				horizon = math.Max(horizon, analysis.signals.siteOpportunity)
				urgency = analysis.signals.builderNeed
			}
			efficiency := costEfficiency(
				validation.CoinCost,
				validation.ResourceCost,
				analysis,
			)
			goalFit := recruitGoalFit(goal, unit)
			safety := localSafety(analysis, position)
			reserve := reservePenalty(validation.ResourceCost, analysis)
			utility := candidateUtility(
				goalFit,
				0,
				horizon,
				urgency,
				efficiency,
				safety,
				novelty(key, recent),
				0.12*reserve,
				0.10*routeValue,
			)
			if unit == game.UnitScout {
				scouts := float64(analysis.unitCounts[game.UnitScout])
				scoutDemand := math.Max(
					analysis.signals.builderNeed,
					analysis.signals.siteOpportunity/(1+1.2*scouts),
				)
				utility *= personality.Economy * (0.20 + 0.80*scoutDemand)
			} else {
				utility *= 0.5*personality.Aggression + 0.5*personality.Defense
			}
			if !add(candidate{
				key:         key,
				description: fmt.Sprintf("Recruit %s at %d,%d", unit, position.X, position.Y),
				manual: &ManualAction{
					Type:    game.ActionRecruit,
					Recruit: &payload,
				},
				utility: utility,
			}) {
				return
			}
		}
	}
}

func generateAttackCandidates(
	world *WorldSnapshot,
	analysis worldAnalysis,
	goal Goal,
	personality Personality,
	recent []string,
	unit unitRef,
	existing map[game.Hex]game.AttackOrder,
	routeValue float64,
	add func(candidate) bool,
) {
	if order, ok := existing[unit.Position]; ok &&
		game.HexAdjacent(order.From, order.TargetTile) {
		return
	}
	stats := game.GetUnitStats(unit.Data.Type)
	for _, position := range neighborHexes(unit.Position) {
		payload := game.AttackActionPayload{From: unit.Position, To: position}
		validation := game.ValidateAdjacentAttackAction(
			&world.Map,
			analysis.owner,
			payload,
		)
		if !validation.Valid {
			continue
		}
		cell := world.Map.GetCell(position)
		hp, destruction := attackTargetValues(cell, analysis.owner)
		kill := 0.0
		if int(stats.Attack) >= hp {
			kill = clamp01(float64(destruction) / 30)
		} else {
			kill = 0.4 * clamp01(float64(stats.Attack)/max(float64(hp), 1))
		}
		targetValue := 0.4
		for _, target := range analysis.targets {
			if target.Position == position {
				targetValue = math.Max(targetValue, target.Score)
			}
		}
		key := fmt.Sprintf("attack:%d:%d:%d:%d", unit.Position.X, unit.Position.Y, position.X, position.Y)
		utility := candidateUtility(
			attackGoalFit(goal),
			kill,
			targetValue,
			math.Max(analysis.signals.townhallThreat, analysis.signals.scorePressure),
			1,
			localSafety(analysis, unit.Position),
			novelty(key, recent),
			0,
			0.08*routeValue,
		)
		utility *= 0.65*personality.Aggression + 0.35*personality.Risk
		if !add(candidate{
			key:         key,
			description: fmt.Sprintf("Attack %d,%d", position.X, position.Y),
			manual: &ManualAction{
				Type:   game.ActionAttack,
				Attack: &payload,
			},
			utility: utility,
		}) {
			return
		}
	}
}

func candidateUtility(
	goalFit float64,
	immediate float64,
	horizon float64,
	urgency float64,
	efficiency float64,
	safety float64,
	noveltyValue float64,
	reservePenalty float64,
	routeDelayPenalty float64,
) float64 {
	return clamp01(
		0.30*clamp01(goalFit) +
			0.18*clamp01(immediate) +
			0.18*clamp01(horizon) +
			0.12*clamp01(urgency) +
			0.10*clamp01(efficiency) +
			0.07*clamp01(safety) +
			0.05*clamp01(noveltyValue) -
			reservePenalty -
			routeDelayPenalty,
	)
}

func routeGoalFit(goal Goal) float64 {
	switch goal {
	case GoalDefend, GoalRaid, GoalConquer, GoalRecover:
		return 0.90
	case GoalExpand, GoalBootstrap:
		return 0.72
	default:
		return 0.60
	}
}

func buildGoalFit(goal Goal, building game.BuildingType) float64 {
	if building == game.BuildingBarracks {
		switch goal {
		case GoalMobilize, GoalDefend:
			return 1
		case GoalRaid, GoalConquer, GoalRecover:
			return 0.78
		default:
			return 0.45
		}
	}
	switch goal {
	case GoalBootstrap, GoalExpand:
		return 1
	case GoalPreserveLead:
		return 0.72
	case GoalMobilize:
		return 0.60
	default:
		return 0.42
	}
}

func recruitGoalFit(goal Goal, unit game.UnitType) float64 {
	if unit == game.UnitScout {
		if goal == GoalBootstrap || goal == GoalExpand {
			return 1
		}
		return 0.35
	}
	switch goal {
	case GoalMobilize, GoalDefend:
		return 1
	case GoalRaid, GoalConquer, GoalRecover:
		return 0.90
	case GoalPreserveLead:
		return 0.72
	default:
		return 0.48
	}
}

func attackGoalFit(goal Goal) float64 {
	switch goal {
	case GoalDefend, GoalRaid, GoalConquer, GoalRecover:
		return 1
	case GoalMobilize:
		return 0.75
	default:
		return 0.45
	}
}

func buildingUrgency(
	analysis worldAnalysis,
	building game.BuildingType,
	tile game.TileType,
) float64 {
	switch building {
	case game.BuildingFarm:
		return analysis.resourceNeed[game.ResourceFood]
	case game.BuildingForester:
		return analysis.resourceNeed[game.ResourceWood]
	case game.BuildingMine:
		switch tile {
		case game.TileRock:
			return analysis.resourceNeed[game.ResourceStone]
		case game.TileIron:
			return analysis.resourceNeed[game.ResourceIron]
		case game.TileGold:
			return math.Max(0.45, analysis.resourceNeed[game.ResourceGold])
		}
	case game.BuildingBarracks:
		return math.Max(analysis.signals.militaryGap, analysis.signals.townhallThreat)
	case game.BuildingBank:
		goldCost := game.BuildingConsumes(game.BuildingBank)[game.ResourceGold]
		if analysis.funds.Resources[game.ResourceGold]+analysis.income[game.ResourceGold] < goldCost {
			return 0.05 * analysis.resourceNeed[game.ResourceGold]
		}
		return math.Max(0.55, analysis.resourceNeed[game.ResourceGold])
	}
	return 0
}

func costEfficiency(coins int32, resources game.Resources, analysis worldAnalysis) float64 {
	coinBurden := float64(coins) / max(float64(analysis.funds.Coins), 1)
	resourceBurden := 0.0
	for resource, amount := range resources {
		available := max(float64(analysis.funds.Resources[resource]), 1)
		resourceBurden += float64(amount) / available *
			(0.5 + 0.5*analysis.resourceNeed[resource])
	}
	return 1 / (1 + 0.5*coinBurden + 0.5*resourceBurden)
}

func reservePenalty(resources game.Resources, analysis worldAnalysis) float64 {
	penalty := 0.0
	for resource, amount := range resources {
		if analysis.funds.Resources[resource] < amount {
			return 1
		}
		remaining := float64(analysis.funds.Resources[resource] - amount)
		if remaining < 4 {
			penalty = math.Max(penalty, analysis.resourceNeed[resource])
		}
	}
	return clamp01(penalty)
}

func localSafety(analysis worldAnalysis, position game.Hex) float64 {
	var pressure float64
	for _, enemy := range analysis.enemyUnits {
		distance := position.Distance(enemy.Position)
		if distance <= 3 {
			pressure += combatPower(enemy.Data) / float64(distance+1)
		}
	}
	return 1 / (1 + pressure/10)
}

func novelty(key string, recent []string) float64 {
	repeats := 0
	for _, previous := range recent {
		if previous == key {
			repeats++
		}
	}
	return clamp01(1 - float64(repeats)/3)
}

func attackTargetValues(cell *game.Cell, owner int8) (hp int, destruction int32) {
	if cell == nil {
		return 1, 0
	}
	for _, unit := range cell.Units {
		if unit.Owner != owner {
			return int(unit.HP), game.UnitDestructionScore(unit.Type)
		}
	}
	if cell.HasBuilding() && cell.Owner >= 0 && cell.Owner != owner {
		return int(cell.Building.HP), game.BuildingDestructionScore(cell.BuildingType(), cell.Tile)
	}
	return 1, 0
}

func neighborHexes(position game.Hex) []game.Hex {
	if position.X&1 == 0 {
		return []game.Hex{
			position.Add(game.NewHex(-1, -1)),
			position.Add(game.NewHex(0, -1)),
			position.Add(game.NewHex(1, -1)),
			position.Add(game.NewHex(1, 0)),
			position.Add(game.NewHex(0, 1)),
			position.Add(game.NewHex(-1, 0)),
		}
	}
	return []game.Hex{
		position.Add(game.NewHex(-1, 0)),
		position.Add(game.NewHex(0, -1)),
		position.Add(game.NewHex(1, 0)),
		position.Add(game.NewHex(1, 1)),
		position.Add(game.NewHex(0, 1)),
		position.Add(game.NewHex(-1, 1)),
	}
}
