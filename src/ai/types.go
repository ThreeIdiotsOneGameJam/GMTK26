package ai

import (
	"fmt"

	"github.com/threeidiotsonegamejam/gmtk26/src/game"
)

type Goal uint8

const (
	GoalBootstrap Goal = iota
	GoalDefend
	GoalExpand
	GoalMobilize
	GoalRaid
	GoalConquer
	GoalPreserveLead
	GoalRecover
)

func (goal Goal) String() string {
	switch goal {
	case GoalBootstrap:
		return "Bootstrap"
	case GoalDefend:
		return "Defend"
	case GoalExpand:
		return "Expand"
	case GoalMobilize:
		return "Mobilize"
	case GoalRaid:
		return "Raid"
	case GoalConquer:
		return "Conquer"
	case GoalPreserveLead:
		return "PreserveLead"
	case GoalRecover:
		return "Recover"
	default:
		return fmt.Sprintf("Goal(%d)", goal)
	}
}

type FactionSnapshot struct {
	Index     int
	Coins     int32
	Points    int32
	Resources game.Resources
	Alive     bool
}

// WorldSnapshot contains only resolved, public gameplay state. In particular
// it deliberately excludes players, AI flags, deadlines, pending actions and
// every faction's private route queues.
type WorldSnapshot struct {
	Round       int32
	TotalRounds int32
	Map         game.Map
	Factions    [4]FactionSnapshot
	Movements   []game.MovementEvent
	Attacks     []game.AttackEvent
}

func NewWorldSnapshot(
	g *game.Game,
	movements []game.MovementEvent,
	attacks []game.AttackEvent,
) WorldSnapshot {
	snapshot := WorldSnapshot{
		TotalRounds: game.TotalMatchRounds,
		Movements:   cloneMovements(movements),
		Attacks:     append([]game.AttackEvent(nil), attacks...),
	}
	if g == nil {
		return snapshot
	}
	snapshot.Round = g.Round
	snapshot.Map = cloneMap(g.Map)
	for i, faction := range g.Factions {
		snapshot.Factions[i] = FactionSnapshot{
			Index:     i,
			Coins:     faction.Coins,
			Points:    faction.Points,
			Resources: cloneResources(faction.Resources),
			Alive:     faction.Alive,
		}
	}
	return snapshot
}

type OwnState struct {
	Owner          int8
	MovementOrders []game.MovementOrder
	AttackOrders   []game.AttackOrder
	LastResult     *game.ActionResult
}

type ManualAction struct {
	Type    game.ActionType
	Build   *game.BuildActionPayload
	Recruit *game.RecruitActionPayload
	Attack  *game.AttackActionPayload
}

type OrderKind uint8

const (
	OrderMove OrderKind = iota
	OrderAttack
)

type OrderCommand struct {
	Kind OrderKind
	From game.Hex
	To   game.Hex
}

type Personality struct {
	Economy     float64
	Expansion   float64
	Defense     float64
	Aggression  float64
	Risk        float64
	Opportunism float64
}

type ScoredAlternative struct {
	Description string
	Utility     float64
}

type DecisionTrace struct {
	Round              int32
	Faction            int8
	Goal               Goal
	GoalUtility        float64
	Choice             string
	ChoiceUtility      float64
	Alternatives       []ScoredAlternative
	Candidates         int
	DetailedCandidates int
	PathQueries        int
	Target             game.Hex
	TargetOwner        int8
	HasTarget          bool
	Personality        Personality
}

type Plan struct {
	Manual *ManualAction
	Orders []OrderCommand
	Trace  DecisionTrace
}

type Planner interface {
	Plan(*WorldSnapshot, OwnState) Plan
}

func cloneMap(source game.Map) game.Map {
	cloned := game.Map{
		Seed:     source.Seed,
		GridSize: source.GridSize,
		Grid:     make([][]game.Cell, len(source.Grid)),
	}
	for x := range source.Grid {
		cloned.Grid[x] = make([]game.Cell, len(source.Grid[x]))
		for y, cell := range source.Grid[x] {
			cloned.Grid[x][y] = cell
			if cell.Building != nil {
				building := *cell.Building
				cloned.Grid[x][y].Building = &building
			}
			cloned.Grid[x][y].Units = append([]game.UnitData(nil), cell.Units...)
		}
	}
	return cloned
}

func cloneResources(source game.Resources) game.Resources {
	cloned := make(game.Resources, len(source))
	for resource, amount := range source {
		cloned[resource] = amount
	}
	return cloned
}

func cloneMovements(source []game.MovementEvent) []game.MovementEvent {
	cloned := make([]game.MovementEvent, len(source))
	for i, movement := range source {
		cloned[i] = movement
		cloned[i].Path = append([]game.Hex(nil), movement.Path...)
	}
	return cloned
}
