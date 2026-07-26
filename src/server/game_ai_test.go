package server

import (
	"strconv"
	"strings"
	"testing"

	gameai "github.com/threeidiotsonegamejam/gmtk26/src/ai"
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
)

type recordingPlanner struct {
	snapshots []*gameai.WorldSnapshot
	owns      []gameai.OwnState
	plan      gameai.Plan
	panic     bool
}

func (planner *recordingPlanner) Plan(
	snapshot *gameai.WorldSnapshot,
	own gameai.OwnState,
) gameai.Plan {
	if planner.panic {
		panic("test panic")
	}
	planner.snapshots = append(planner.snapshots, snapshot)
	planner.owns = append(planner.owns, own)
	return planner.plan
}

func TestAIResolvesOpeningThroughAuthoritativeRecruitment(t *testing.T) {
	g := actionTestGame(3, 3)
	for i := range g.Factions {
		g.Factions[i].Alive = false
		g.Factions[i].AI = false
		g.Factions[i].Coins = game.StartingCoins
		g.Factions[i].Resources = make(game.Resources)
	}
	g.Factions[0].Alive = true
	g.Factions[0].AI = true
	townhall := g.Map.GetCell(game.NewHex(1, 1))
	townhall.Owner = 0
	townhall.Building = &game.BuildingData{
		Type: game.BuildingTownhall,
		HP:   game.BuildingMaxHP(game.BuildingTownhall),
	}
	gi := NewGameInstance(1, g, nil)
	gi.ensureAIControllersLocked()

	gi.resolveRoundLocked()

	scouts := 0
	for x := range g.Map.Grid {
		for y := range g.Map.Grid[x] {
			cell := &g.Map.Grid[x][y]
			if cell.HasUnits() &&
				cell.Units[0].Owner == 0 &&
				cell.Units[0].Type == game.UnitScout {
				scouts++
			}
		}
	}
	if scouts != 1 {
		t.Fatalf("Scout count = %d, want 1", scouts)
	}
	if g.Factions[0].Coins != game.StartingCoins+1-game.UnitCost(game.UnitScout) {
		t.Fatalf("coins = %d after authoritative income and recruitment", g.Factions[0].Coins)
	}
}

func TestAllAIPlannersReceiveSamePublicSnapshot(t *testing.T) {
	g := actionTestGame(2, 1)
	g.Factions[0].AI = true
	g.Factions[1].AI = true
	first := &recordingPlanner{}
	second := &recordingPlanner{}
	gi := NewGameInstance(1, g, nil)
	gi.aiControllers[0] = first
	gi.aiControllers[1] = second
	gi.actions[2] = &submittedAction{Type: game.ActionPass}
	gi.movementOrders[2] = []game.MovementOrder{{
		Current: game.NewHex(0, 0), Destination: game.NewHex(1, 0),
	}}

	gi.planAIActionsLocked()

	if len(first.snapshots) != 1 || len(second.snapshots) != 1 {
		t.Fatal("planners were not called exactly once")
	}
	if first.snapshots[0] != second.snapshots[0] {
		t.Fatal("AI planners did not share one immutable round snapshot")
	}
	if len(first.owns[0].MovementOrders) != 0 ||
		len(second.owns[0].MovementOrders) != 0 {
		t.Fatal("AI received another faction's private movement orders")
	}
}

func TestInvalidAIOrderIsRejected(t *testing.T) {
	g := actionTestGame(3, 1)
	g.Factions[0].AI = true
	putUnit(g, game.NewHex(0, 0), game.UnitScout, 0)
	gi := NewGameInstance(1, g, nil)

	gi.applyAIPlanLocked(0, gameai.Plan{
		Orders: []gameai.OrderCommand{{
			Kind: gameai.OrderAttack,
			From: game.NewHex(0, 0),
			To:   game.NewHex(2, 0),
		}},
	})

	if len(gi.attackOrders[0]) != 0 {
		t.Fatalf("invalid Scout attack order was accepted: %v", gi.attackOrders[0])
	}
}

func TestDisconnectSchedulesAndActivatesAITakeover(t *testing.T) {
	g := actionTestGame(2, 1)
	human := NewClient(nil)
	other := NewClient(nil)
	gi := NewGameInstance(1, g, []*Client{human, other})
	human.JoinGame(gi)
	other.JoinGame(gi)
	gi.actions[0] = &submittedAction{Type: game.ActionPass}
	gi.movementOrders[0] = []game.MovementOrder{{
		Current: game.NewHex(0, 0), Destination: game.NewHex(1, 0),
	}}
	human.LeaveGame()

	if !gi.pendingAITakeovers[0] {
		t.Fatal("disconnect did not schedule AI takeover")
	}
	gi.activateAITakeoversLocked()

	if !g.Factions[0].AI || gi.aiControllers[0] == nil {
		t.Fatal("disconnected faction did not receive AI controller")
	}
	if _, exists := gi.actions[0]; exists {
		t.Fatal("stale pending human action survived takeover")
	}
	if len(gi.movementOrders[0]) != 1 {
		t.Fatal("takeover discarded persistent movement orders")
	}
	if !gi.hasConnectedPlayers() {
		t.Fatal("remaining human was not kept connected")
	}
}

func TestLastHumanDisconnectStillLeavesNoConnectedPlayers(t *testing.T) {
	g := actionTestGame(1, 1)
	human := NewClient(nil)
	gi := NewGameInstance(1, g, []*Client{human})
	human.JoinGame(gi)

	human.LeaveGame()

	if gi.hasConnectedPlayers() {
		t.Fatal("last human disconnect did not terminate connected-player state")
	}
}

func TestPlannerPanicIsIsolated(t *testing.T) {
	g := actionTestGame(1, 1)
	planner := &recordingPlanner{panic: true}
	snapshot := gameai.NewWorldSnapshot(g, nil, nil)

	if _, ok := safelyPlanAI(planner, &snapshot, gameai.OwnState{Owner: 0}); ok {
		t.Fatal("panicking planner reported success")
	}
}

func TestAISelfPlayMaintainsValidEconomyAndActionVariety(t *testing.T) {
	for _, seed := range []int64{1717, 2026, 9091} {
		t.Run(strconv.FormatInt(seed, 10), func(t *testing.T) {
			g := &game.Game{Round: 1}
			g.Map.Seed = seed
			g.Map.GridSize.X = 24
			g.Map.GridSize.Y = 24
			g.Map.Generate()
			for i := range g.Factions {
				g.Factions[i] = game.Faction{
					Index:     i,
					AI:        true,
					Coins:     game.StartingCoins,
					Resources: make(game.Resources),
					Alive:     true,
				}
			}
			gi := NewGameInstance(1, g, make([]*Client, len(g.Factions)))
			gi.assignStartingCells()
			gi.ensureAIControllersLocked()

			activity := 0
			attacks := 0
			actionKinds := make(map[string]int)
			for range game.TotalMatchRounds {
				beforeBuildings, beforeUnits := objectCounts(&g.Map)
				gi.resolveRoundLocked()
				afterBuildings, afterUnits := objectCounts(&g.Map)
				if beforeBuildings != afterBuildings ||
					beforeUnits != afterUnits ||
					len(gi.movementEvents) > 0 ||
					len(gi.attackEvents) > 0 {
					activity++
				}
				attacks += len(gi.attackEvents)
				for i, faction := range g.Factions {
					if faction.Coins < 0 {
						t.Fatalf("faction %d has negative coins", i)
					}
				}
				for _, trace := range gi.aiTraces {
					kind := strings.Fields(trace.Choice)
					if len(kind) > 0 {
						actionKinds[kind[0]]++
					}
				}
			}
			if activity < 30 {
				t.Fatalf("AI self-play produced too little activity: %d active rounds", activity)
			}
			if actionKinds["Build"] == 0 ||
				actionKinds["Recruit"] == 0 ||
				actionKinds["Advance"] == 0 ||
				actionKinds["Attack"] == 0 ||
				attacks == 0 {
				t.Fatalf(
					"AI self-play lacked core action variety: actions=%v attacks=%d",
					actionKinds,
					attacks,
				)
			}
			buildingTypes, unitTypes := objectTypeCounts(&g.Map)
			if buildingTypes[game.BuildingBarracks] == 0 {
				t.Fatal("AI never established military production")
			}
			if unitTypes[game.UnitPeasant]+
				unitTypes[game.UnitArcher]+
				unitTypes[game.UnitKnight] == 0 {
				t.Fatal("AI never retained a combat unit")
			}
			t.Logf(
				"activity=%d attacks=%d actions=%v",
				activity,
				attacks,
				actionKinds,
			)
		})
	}
}

func objectTypeCounts(m *game.Map) (map[game.BuildingType]int, map[game.UnitType]int) {
	buildings := make(map[game.BuildingType]int)
	units := make(map[game.UnitType]int)
	for x := range m.Grid {
		for y := range m.Grid[x] {
			cell := &m.Grid[x][y]
			if cell.HasBuilding() {
				buildings[cell.BuildingType()]++
			}
			for _, unit := range cell.Units {
				units[unit.Type]++
			}
		}
	}
	return buildings, units
}

func objectCounts(m *game.Map) (buildings, units int) {
	for x := range m.Grid {
		for y := range m.Grid[x] {
			cell := &m.Grid[x][y]
			if cell.HasBuilding() {
				buildings++
			}
			units += len(cell.Units)
		}
	}
	return buildings, units
}
