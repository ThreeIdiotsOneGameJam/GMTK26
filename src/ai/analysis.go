package ai

import (
	"math"
	"sort"

	"github.com/threeidiotsonegamejam/gmtk26/src/game"
)

type unitRef struct {
	Position game.Hex
	Data     game.UnitData
}

type buildingRef struct {
	Position game.Hex
	Owner    int8
	Tile     game.TileType
	Data     game.BuildingData
}

type siteRef struct {
	Position game.Hex
	Building game.BuildingType
	Score    float64
}

type targetRef struct {
	Position game.Hex
	Owner    int8
	Score    float64
}

type opponentMemory struct {
	Retaliation   float64
	ScoreVelocity float64
	LastPoints    int32
}

type signals struct {
	builderNeed            float64
	productionGap          float64
	siteOpportunity        float64
	townhallThreat         float64
	militaryGap            float64
	militaryAdvantage      float64
	raidOpportunity        float64
	scorePressure          float64
	leadSecurity           float64
	targetVulnerability    float64
	routeValue             float64
	controlOpportunity     float64
	observedAggression     float64
	destructionOpportunity float64
}

type worldAnalysis struct {
	owner           int8
	faction         game.Faction
	funds           game.Funds
	forecast        game.Funds
	income          game.Resources
	progress        float64
	remainingRounds int32
	remainingPulses int32

	townhall       game.Hex
	hasTownhall    bool
	ownUnits       []unitRef
	ownBuildings   []buildingRef
	enemyUnits     []unitRef
	enemyBuildings []buildingRef
	sites          []siteRef
	targets        []targetRef

	unitCounts     map[game.UnitType]int
	buildingCounts map[game.BuildingType]int
	resourceNeed   map[game.ResourceType]float64
	ownPower       float64
	maxEnemyPower  float64
	signals        signals
}

func analyzeWorld(
	snapshot *WorldSnapshot,
	own OwnState,
	memory [4]opponentMemory,
	forecastRounds int32,
) worldAnalysis {
	analysis := worldAnalysis{
		owner:           own.Owner,
		remainingRounds: max(int32(0), snapshot.TotalRounds-snapshot.Round+1),
		unitCounts:      make(map[game.UnitType]int),
		buildingCounts:  make(map[game.BuildingType]int),
		resourceNeed:    make(map[game.ResourceType]float64),
		income:          make(game.Resources),
	}
	if snapshot.TotalRounds > 0 {
		analysis.progress = clamp01(float64(snapshot.Round-1) / float64(snapshot.TotalRounds))
	}
	if int(own.Owner) >= 0 && int(own.Owner) < len(snapshot.Factions) {
		faction := snapshot.Factions[own.Owner]
		analysis.faction = game.Faction{
			Index:     faction.Index,
			Coins:     faction.Coins,
			Points:    faction.Points,
			Resources: cloneResources(faction.Resources),
			Alive:     faction.Alive,
		}
		analysis.funds = game.ProjectedRoundFunds(&snapshot.Map, own.Owner, analysis.faction)
		horizon := min(
			max(forecastRounds, int32(1)),
			max(analysis.remainingRounds, int32(1)),
		)
		analysis.forecast = game.ProjectedFundsAfterRounds(
			&snapshot.Map,
			own.Owner,
			analysis.faction,
			horizon,
		)
	}
	_, analysis.income = game.FactionRoundIncome(&snapshot.Map, own.Owner)

	powers := [4]float64{}
	for x := range snapshot.Map.Grid {
		for y := range snapshot.Map.Grid[x] {
			position := game.NewHex(int32(x), int32(y))
			cell := &snapshot.Map.Grid[x][y]
			for _, unit := range cell.Units {
				ref := unitRef{Position: position, Data: unit}
				if validOwner(unit.Owner) {
					powers[unit.Owner] += combatPower(unit)
				}
				if unit.Owner == own.Owner {
					analysis.ownUnits = append(analysis.ownUnits, ref)
					analysis.unitCounts[unit.Type]++
				} else if validOwner(unit.Owner) && snapshot.Factions[unit.Owner].Alive {
					analysis.enemyUnits = append(analysis.enemyUnits, ref)
				}
			}
			if !cell.HasBuilding() || !validOwner(cell.Owner) {
				continue
			}
			ref := buildingRef{
				Position: position,
				Owner:    cell.Owner,
				Tile:     cell.Tile,
				Data:     *cell.Building,
			}
			if cell.Owner == own.Owner {
				analysis.ownBuildings = append(analysis.ownBuildings, ref)
				analysis.buildingCounts[cell.BuildingType()]++
				if cell.BuildingType() == game.BuildingTownhall {
					analysis.townhall = position
					analysis.hasTownhall = true
				}
			} else if snapshot.Factions[cell.Owner].Alive {
				analysis.enemyBuildings = append(analysis.enemyBuildings, ref)
			}
		}
	}

	analysis.ownPower = powers[own.Owner]
	for i, power := range powers {
		if int(own.Owner) != i && snapshot.Factions[i].Alive {
			analysis.maxEnemyPower = math.Max(analysis.maxEnemyPower, power)
		}
	}
	analysis.signals.militaryGap = ramp(
		analysis.maxEnemyPower/(analysis.ownPower+4),
		0.9,
		1.8,
	)
	if analysis.ownPower == 0 && len(analysis.enemyBuildings) > 0 {
		analysis.signals.militaryGap = math.Max(analysis.signals.militaryGap, 0.70)
	}
	analysis.signals.militaryAdvantage = clamp01(
		(analysis.ownPower + 4) / (analysis.maxEnemyPower + 4) / 1.5,
	)
	analysis.signals.builderNeed = clamp01(1 - float64(analysis.unitCounts[game.UnitScout])*0.65)

	analysis.calculateScoreSignals(snapshot)
	analysis.calculateThreatSignals(memory)
	analysis.calculateResourceNeed()
	analysis.collectSites(snapshot)
	analysis.collectTargets(snapshot, memory)
	analysis.calculateRouteValue(own)
	return analysis
}

func (analysis *worldAnalysis) calculateScoreSignals(snapshot *WorldSnapshot) {
	ownPoints := analysis.faction.Points
	maxPoints := ownPoints
	minPoints := ownPoints
	living := 0
	for _, faction := range snapshot.Factions {
		if !faction.Alive {
			continue
		}
		living++
		maxPoints = max(maxPoints, faction.Points)
		minPoints = min(minPoints, faction.Points)
	}
	spread := max(float64(maxPoints-minPoints), 10)
	deficit := float64(maxPoints-ownPoints) / spread
	lead := float64(ownPoints-minPoints) / spread
	analysis.signals.scorePressure = clamp01(deficit * analysis.progress)
	if living > 1 && ownPoints == maxPoints {
		analysis.signals.leadSecurity = clamp01(lead * analysis.progress)
	}
	nextPulse := game.ScoreIntervalRounds - snapshot.Round%game.ScoreIntervalRounds
	if nextPulse == game.ScoreIntervalRounds {
		nextPulse = 0
	}
	if nextPulse <= analysis.remainingRounds {
		analysis.remainingPulses = 1 + (analysis.remainingRounds-nextPulse)/game.ScoreIntervalRounds
	}
}

func (analysis *worldAnalysis) calculateThreatSignals(memory [4]opponentMemory) {
	if !analysis.hasTownhall {
		analysis.signals.townhallThreat = 1
		return
	}
	var enemyPressure, friendlyResponse float64
	for _, unit := range analysis.enemyUnits {
		distance := analysis.townhall.Distance(unit.Position)
		if distance <= 10 {
			enemyPressure += combatPower(unit.Data) / float64(distance+1)
		}
	}
	for _, unit := range analysis.ownUnits {
		distance := analysis.townhall.Distance(unit.Position)
		if distance <= 8 {
			friendlyResponse += combatPower(unit.Data) / float64(distance+1)
		}
	}
	ratio := enemyPressure / (friendlyResponse + 2)
	analysis.signals.townhallThreat = ramp(ratio, 0.35, 1.25)
	for _, building := range analysis.ownBuildings {
		if building.Data.Type != game.BuildingTownhall {
			continue
		}
		damage := 1 - float64(building.Data.HP)/float64(game.BuildingMaxHP(game.BuildingTownhall))
		analysis.signals.townhallThreat = clamp01(
			analysis.signals.townhallThreat + 0.35*damage,
		)
	}
	for i, opponent := range memory {
		if int(analysis.owner) == i {
			continue
		}
		analysis.signals.observedAggression = math.Max(
			analysis.signals.observedAggression,
			opponent.Retaliation,
		)
	}
}

func (analysis *worldAnalysis) calculateResourceNeed() {
	militaryDemand := 0.35 + 0.65*math.Max(
		analysis.signals.militaryGap,
		analysis.signals.scorePressure,
	)
	if analysis.signals.townhallThreat > militaryDemand {
		militaryDemand = analysis.signals.townhallThreat
	}
	hasBarracks := analysis.buildingCounts[game.BuildingBarracks] > 0

	desired := map[game.ResourceType]float64{
		game.ResourceFood:  4 + 8*militaryDemand,
		game.ResourceWood:  4 + 6*militaryDemand,
		game.ResourceStone: 0,
		game.ResourceIron:  2 + 4*militaryDemand,
	}
	if !hasBarracks {
		desired[game.ResourceWood] += 6
		desired[game.ResourceStone] += 4
	}
	for resource, amount := range desired {
		projected := float64(analysis.forecast.Resources[resource])
		if amount > 0 {
			analysis.resourceNeed[resource] = clamp01((amount - projected) / amount)
		}
		analysis.signals.productionGap = math.Max(
			analysis.signals.productionGap,
			analysis.resourceNeed[resource],
		)
	}
	desiredCoins := float64(game.BuildingCost(game.BuildingBarracks))
	if hasBarracks {
		desiredCoins = float64(game.UnitCost(game.UnitArcher))
	}
	projectedCoins := float64(analysis.forecast.Coins)
	coinNeed := clamp01((desiredCoins - projectedCoins) / max(desiredCoins, 1))
	analysis.resourceNeed[game.ResourceGold] = coinNeed
	analysis.signals.productionGap = math.Max(analysis.signals.productionGap, coinNeed)
}

func (analysis *worldAnalysis) collectSites(snapshot *WorldSnapshot) {
	buildings := [...]game.BuildingType{
		game.BuildingFarm,
		game.BuildingForester,
		game.BuildingMine,
		game.BuildingBarracks,
		game.BuildingBank,
	}
	for x := range snapshot.Map.Grid {
		for y := range snapshot.Map.Grid[x] {
			position := game.NewHex(int32(x), int32(y))
			cell := snapshot.Map.GetCell(position)
			if cell == nil ||
				cell.Owner != -1 && cell.Owner != analysis.owner ||
				cell.HasUnits() && cell.Units[0].Owner != analysis.owner {
				continue
			}
			for _, building := range buildings {
				if !game.BuildingCanPlace(&snapshot.Map, building, position) {
					continue
				}
				score := analysis.siteScore(position, building, cell.Tile)
				analysis.sites = append(analysis.sites, siteRef{
					Position: position,
					Building: building,
					Score:    score,
				})
				analysis.signals.siteOpportunity = math.Max(
					analysis.signals.siteOpportunity,
					score,
				)
				control := float64(game.BuildingControlScore(building, cell.Tile)) / 5
				analysis.signals.controlOpportunity = math.Max(
					analysis.signals.controlOpportunity,
					control*clamp01(float64(analysis.remainingPulses)/5),
				)
			}
		}
	}
	sort.SliceStable(analysis.sites, func(i, j int) bool {
		if analysis.sites[i].Score == analysis.sites[j].Score {
			if analysis.sites[i].Position.X == analysis.sites[j].Position.X {
				if analysis.sites[i].Position.Y == analysis.sites[j].Position.Y {
					return analysis.sites[i].Building < analysis.sites[j].Building
				}
				return analysis.sites[i].Position.Y < analysis.sites[j].Position.Y
			}
			return analysis.sites[i].Position.X < analysis.sites[j].Position.X
		}
		return analysis.sites[i].Score > analysis.sites[j].Score
	})
}

func (analysis *worldAnalysis) siteScore(
	position game.Hex,
	building game.BuildingType,
	tile game.TileType,
) float64 {
	need := 0.35
	incomeValue := 0.0
	switch building {
	case game.BuildingFarm:
		need = analysis.resourceNeed[game.ResourceFood]
		incomeValue = 0.55
	case game.BuildingForester:
		need = analysis.resourceNeed[game.ResourceWood]
		incomeValue = 0.55
	case game.BuildingMine:
		switch tile {
		case game.TileRock:
			need = analysis.resourceNeed[game.ResourceStone]
			incomeValue = 0.55
		case game.TileIron:
			need = analysis.resourceNeed[game.ResourceIron]
			incomeValue = 0.55
		case game.TileGold:
			need = math.Max(analysis.resourceNeed[game.ResourceGold], 0.45)
			incomeValue = 1
		}
	case game.BuildingBarracks:
		need = math.Max(analysis.signals.militaryGap, analysis.signals.townhallThreat)
		if analysis.buildingCounts[game.BuildingBarracks] == 0 {
			need = math.Max(need, 0.55)
		} else {
			need *= 0.45
		}
	case game.BuildingBank:
		availableGold := analysis.forecast.Resources[game.ResourceGold]
		need = analysis.resourceNeed[game.ResourceGold]
		if availableGold >= game.BuildingConsumes(game.BuildingBank)[game.ResourceGold] {
			need = math.Max(need, 0.55)
			incomeValue = 0.85
		} else {
			need *= 0.15
			incomeValue = 0.05
		}
	}
	control := float64(game.BuildingControlScore(building, tile)) / 5
	reachability := 0.5
	defensibility := 0.5
	if analysis.hasTownhall {
		distance := float64(analysis.townhall.Distance(position))
		reachability = 1 / (1 + distance/8)
		defensibility = 1 / (1 + distance/10)
	}
	timeReturn := clamp01(float64(analysis.remainingRounds) / 30)
	saturation := clamp01(float64(analysis.buildingCounts[building]-1) * 0.12)
	return clamp01(
		0.30*need +
			0.25*incomeValue*timeReturn +
			0.20*control +
			0.15*reachability +
			0.10*defensibility -
			saturation,
	)
}

func (analysis *worldAnalysis) collectTargets(
	snapshot *WorldSnapshot,
	memory [4]opponentMemory,
) {
	for _, building := range analysis.enemyBuildings {
		cell := snapshot.Map.GetCell(building.Position)
		if cell == nil {
			continue
		}
		score := analysis.targetScore(
			building.Position,
			building.Owner,
			game.BuildingDestructionScore(building.Data.Type, building.Tile),
			int(building.Data.HP),
			int(game.BuildingMaxHP(building.Data.Type)),
			building.Data.Type == game.BuildingTownhall ||
				building.Data.Type == game.BuildingBarracks ||
				building.Data.Type == game.BuildingMine && building.Tile == game.TileGold,
			memory,
			snapshot,
		)
		analysis.targets = append(analysis.targets, targetRef{
			Position: building.Position,
			Owner:    building.Owner,
			Score:    score,
		})
	}
	for _, unit := range analysis.enemyUnits {
		score := analysis.targetScore(
			unit.Position,
			unit.Data.Owner,
			game.UnitDestructionScore(unit.Data.Type),
			int(unit.Data.HP),
			int(game.GetUnitStats(unit.Data.Type).MaxHP),
			false,
			memory,
			snapshot,
		)
		analysis.targets = append(analysis.targets, targetRef{
			Position: unit.Position,
			Owner:    unit.Data.Owner,
			Score:    score,
		})
	}
	sort.SliceStable(analysis.targets, func(i, j int) bool {
		if analysis.targets[i].Score == analysis.targets[j].Score {
			if analysis.targets[i].Position.X == analysis.targets[j].Position.X {
				return analysis.targets[i].Position.Y < analysis.targets[j].Position.Y
			}
			return analysis.targets[i].Position.X < analysis.targets[j].Position.X
		}
		return analysis.targets[i].Score > analysis.targets[j].Score
	})
	if len(analysis.targets) > 0 {
		analysis.signals.raidOpportunity = analysis.targets[0].Score
		analysis.signals.targetVulnerability = analysis.targets[0].Score
		analysis.signals.destructionOpportunity = analysis.targets[0].Score
	}
}

func (analysis *worldAnalysis) targetScore(
	position game.Hex,
	owner int8,
	destruction int32,
	hp int,
	maxHP int,
	strategic bool,
	memory [4]opponentMemory,
	snapshot *WorldSnapshot,
) float64 {
	destructionValue := clamp01(float64(destruction) / 30)
	strategicValue := 0.25
	if strategic {
		strategicValue = 1
	}
	vulnerability := 1 - clamp01(float64(hp)/max(float64(maxHP), 1))
	vulnerability = math.Max(vulnerability, 0.20)
	proximity := 0.25
	if analysis.hasTownhall {
		proximity = 1 / (1 + float64(analysis.townhall.Distance(position))/10)
	}
	maxPoints := int32(0)
	minPoints := int32(math.MaxInt32)
	for _, faction := range snapshot.Factions {
		if faction.Alive {
			maxPoints = max(maxPoints, faction.Points)
			minPoints = min(minPoints, faction.Points)
		}
	}
	threat := 0.0
	underdogPenalty := 0.0
	if validOwner(owner) {
		spread := max(float64(maxPoints-minPoints), 10)
		threat = clamp01(float64(snapshot.Factions[owner].Points-minPoints) / spread)
		threat = clamp01(
			threat + 0.25*clamp01(memory[owner].ScoreVelocity/10),
		)
		if snapshot.Factions[owner].Points == minPoints &&
			analysis.signals.townhallThreat < 0.5 {
			underdogPenalty = 0.10
		}
	}
	retaliation := 0.0
	if validOwner(owner) {
		retaliation = memory[owner].Retaliation
	}
	localRisk := 0.0
	for _, unit := range analysis.enemyUnits {
		if unit.Position.Distance(position) <= 3 {
			localRisk += unitPower(unit.Data) / 40
		}
	}
	return clamp01(
		0.30*destructionValue +
			0.20*strategicValue +
			0.20*vulnerability +
			0.15*proximity +
			0.10*threat +
			0.05*retaliation -
			0.15*clamp01(localRisk) -
			underdogPenalty,
	)
}

func (analysis *worldAnalysis) calculateRouteValue(own OwnState) {
	if len(own.MovementOrders)+len(own.AttackOrders) == 0 {
		return
	}
	best := 0.0
	for _, order := range own.MovementOrders {
		distance := float64(order.Current.Distance(order.Destination))
		best = math.Max(best, 0.45+0.35/(1+distance/5))
	}
	for _, order := range own.AttackOrders {
		distance := float64(order.From.Distance(order.TargetTile))
		value := 0.50 + 0.40/(1+distance/5)
		for _, target := range analysis.targets {
			if target.Position == order.TargetTile {
				value = math.Max(value, target.Score)
				break
			}
		}
		best = math.Max(best, value)
	}
	analysis.signals.routeValue = clamp01(best)
}

func unitPower(unit game.UnitData) float64 {
	stats := game.GetUnitStats(unit.Type)
	return float64(unit.HP) +
		2*float64(stats.Attack) +
		0.5*float64(game.UnitMovementBudget(unit.Type))
}

func combatPower(unit game.UnitData) float64 {
	if game.GetUnitStats(unit.Type).Attack <= 0 {
		return 0
	}
	return unitPower(unit)
}

func validOwner(owner int8) bool {
	return owner >= 0 && owner < 4
}
