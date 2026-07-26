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

func TestAIExactMineSpendUsesAuthoritativeIncome(t *testing.T) {
	g := actionTestGame(2, 1)
	for i := range g.Factions {
		g.Factions[i].Alive = false
		g.Factions[i].AI = false
	}
	g.Factions[0] = game.Faction{
		Index:     0,
		AI:        true,
		Alive:     true,
		Coins:     game.BuildingCost(game.BuildingMine) - 1,
		Resources: make(game.Resources),
	}
	source := game.NewHex(0, 0)
	target := game.NewHex(1, 0)
	g.Map.GetCell(source).Owner = 0
	g.Map.GetCell(source).Building = &game.BuildingData{
		Type: game.BuildingTownhall,
		HP:   game.BuildingMaxHP(game.BuildingTownhall),
	}
	putUnit(g, source, game.UnitScout, 0)
	g.Map.GetCell(target).Tile = game.TileRock
	planner := &recordingPlanner{plan: gameai.Plan{
		Manual: &gameai.ManualAction{
			Type: game.ActionBuild,
			Build: &game.BuildActionPayload{
				From:     source,
				To:       target,
				Building: game.BuildingMine,
			},
		},
	}}
	gi := NewGameInstance(1, g, nil)
	gi.aiControllers[0] = planner

	gi.resolveRoundLocked()

	if got := g.Factions[0].Coins; got != 0 {
		t.Fatalf("AI Coins = %d, want exact income then Mine spend to leave 0", got)
	}
	if got := g.Map.GetCell(target).BuildingType(); got != game.BuildingMine {
		t.Fatalf("AI build = %s, want Mine", got)
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
				actionKinds["Attack"] < 20 ||
				attacks < 40 {
				t.Fatalf(
					"AI self-play lacked sustained PvP pressure: actions=%v attacks=%d",
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

func TestAISelfPlayConservesAuthoritativeFundsEveryRound(t *testing.T) {
	g := &game.Game{Round: 1}
	g.Map.Seed = 6060
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

	for range int32(90) {
		gi.planAIActionsLocked()
		planned := make(map[int]*submittedAction, len(gi.actions))
		for factionIdx, action := range gi.actions {
			planned[factionIdx] = action
		}

		var expected [4]game.Faction
		var startedAlive [4]bool
		for i, faction := range g.Factions {
			startedAlive[i] = faction.Alive
			if !faction.Alive {
				continue
			}
			projected := game.ProjectedRoundFunds(&g.Map, int8(i), faction)
			expected[i] = game.Faction{
				Coins:     projected.Coins,
				Resources: projected.Resources,
			}
		}

		gi.processAutoActions()
		gi.processClientActions()

		for i := range g.Factions {
			if !startedAlive[i] {
				continue
			}
			result := gi.actionResults[i]
			action := planned[i]
			if result != nil &&
				result.Status == game.ActionResultSucceeded &&
				!result.Automatic &&
				action != nil {
				coinCost, resourceCost := actionCosts(action)
				if !game.SpendFunds(&expected[i], coinCost, resourceCost) {
					t.Fatalf(
						"round %d faction %d succeeded with unaffordable action %+v",
						g.Round,
						i,
						action,
					)
				}
			}
			if got, want := g.Factions[i].Coins, expected[i].Coins; got != want {
				t.Fatalf(
					"round %d faction %d Coins = %d, ledger wants %d",
					g.Round,
					i,
					got,
					want,
				)
			}
			for resource := game.ResourceUnknown; resource <= game.ResourceFood; resource++ {
				if got, want := g.Factions[i].Resources[resource], expected[i].Resources[resource]; got != want {
					t.Fatalf(
						"round %d faction %d %s = %d, ledger wants %d",
						g.Round,
						i,
						resource,
						got,
						want,
					)
				}
			}
		}

		gi.checkAlive()
		if controlScoreDue(g.Round) {
			gi.awardControlScore()
		}
		g.Round++
		gi.actions = make(map[int]*submittedAction)
	}
}

func actionCosts(action *submittedAction) (int32, game.Resources) {
	switch action.Type {
	case game.ActionBuild:
		if action.Build != nil {
			return game.BuildingCost(action.Build.Building),
				game.BuildingResourceCost(action.Build.Building)
		}
	case game.ActionRecruit:
		if action.Recruit != nil {
			return game.UnitCost(action.Recruit.Unit),
				game.UnitResourceCost(action.Recruit.Unit)
		}
	}
	return 0, make(game.Resources)
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
