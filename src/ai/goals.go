package ai

import (
	"math"

	"github.com/threeidiotsonegamejam/gmtk26/src/game"
)

func goalScores(analysis worldAnalysis, personality Personality) [8]float64 {
	s := analysis.signals
	early := 1 - analysis.progress
	late := analysis.progress
	missingBarracks := 0.0
	if analysis.buildingCounts[game.BuildingBarracks] == 0 {
		missingBarracks = 1
	}

	var scores [8]float64
	scores[GoalBootstrap] = clamp01(
		0.40*s.builderNeed+
			0.35*s.productionGap+
			0.15*s.militaryGap+
			0.10*early,
	) * personality.Economy
	scores[GoalDefend] = clamp01(
		0.65*s.townhallThreat+
			0.20*townhallDamage(analysis)+
			0.15*s.observedAggression,
	) * personality.Defense
	scores[GoalExpand] = clamp01(
		0.50*s.siteOpportunity+
			0.30*s.productionGap+
			0.20*s.controlOpportunity,
	) * personality.Expansion
	scores[GoalMobilize] = clamp01(
		0.50*s.militaryGap+
			0.20*s.townhallThreat+
			0.20*s.raidOpportunity+
			0.10*missingBarracks,
	) * (0.5*personality.Aggression + 0.5*personality.Defense)
	scores[GoalRaid] = clamp01(
		0.50*s.raidOpportunity+
			0.25*s.militaryAdvantage+
			0.25*s.scorePressure,
	) * personality.Opportunism
	scores[GoalConquer] = clamp01(
		0.40*s.targetVulnerability+
			0.30*s.militaryAdvantage+
			0.30*late,
	) * personality.Aggression
	scores[GoalPreserveLead] = clamp01(
		0.45*s.leadSecurity+
			0.30*s.townhallThreat+
			0.25*s.controlOpportunity,
	) * personality.Defense
	scores[GoalRecover] = clamp01(
		0.50*s.scorePressure+
			0.30*s.destructionOpportunity+
			0.20*late,
	) * (0.5*personality.Aggression + 0.5*personality.Risk)

	for i := range scores {
		scores[i] = clamp01(scores[i])
	}
	return scores
}

func townhallDamage(analysis worldAnalysis) float64 {
	for _, building := range analysis.ownBuildings {
		if building.Position != analysis.townhall {
			continue
		}
		maxHP := float64(game.BuildingMaxHP(game.BuildingTownhall))
		return clamp01(1 - float64(building.Data.HP)/maxHP)
	}
	if !analysis.hasTownhall {
		return 1
	}
	return 0
}

func highestGoal(scores [8]float64) (Goal, float64) {
	best := GoalBootstrap
	bestScore := math.Inf(-1)
	for goal := GoalBootstrap; goal <= GoalRecover; goal++ {
		if scores[goal] > bestScore {
			best = goal
			bestScore = scores[goal]
		}
	}
	return best, bestScore
}
