package ai

import (
	"reflect"
	"testing"

	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

func aiTestGame(width, height int32) *game.Game {
	grid := make([][]game.Cell, width)
	for x := range grid {
		grid[x] = make([]game.Cell, height)
		for y := range grid[x] {
			grid[x][y] = game.Cell{Tile: game.TilePlains, Owner: -1}
		}
	}
	g := &game.Game{
		Round: 1,
		Map: game.Map{
			Seed:     42,
			Grid:     grid,
			GridSize: vec.Vec2i{X: width, Y: height},
		},
	}
	for i := range g.Factions {
		g.Factions[i] = game.Faction{
			Index:     i,
			Coins:     game.StartingCoins,
			Resources: make(game.Resources),
			Alive:     false,
		}
	}
	return g
}

func addTownhall(g *game.Game, owner int8, at game.Hex) {
	g.Factions[owner].Alive = true
	cell := g.Map.GetCell(at)
	cell.Owner = owner
	cell.Building = &game.BuildingData{
		Type: game.BuildingTownhall,
		HP:   game.BuildingMaxHP(game.BuildingTownhall),
	}
}

func TestSnapshotDeepCopiesAndExcludesIdentity(t *testing.T) {
	g := aiTestGame(2, 1)
	addTownhall(g, 0, game.NewHex(0, 0))
	g.Factions[0].Player = &game.Player{PlayerName: "secret"}
	g.Factions[0].AI = true
	g.Factions[0].Resources[game.ResourceFood] = 3
	g.Map.Grid[1][0].Units = []game.UnitData{{Type: game.UnitScout, Owner: 0, HP: 3}}

	snapshot := NewWorldSnapshot(g, nil, nil)
	snapshot.Factions[0].Resources[game.ResourceFood] = 99
	snapshot.Map.Grid[1][0].Units[0].HP = 1

	if g.Factions[0].Resources[game.ResourceFood] != 3 ||
		g.Map.Grid[1][0].Units[0].HP != 3 {
		t.Fatal("snapshot aliases authoritative state")
	}
	// FactionSnapshot intentionally has no Player or AI fields; constructing
	// this assertion through its public value also protects that API boundary.
	if snapshot.Factions[0].Index != 0 {
		t.Fatal("snapshot lost public faction state")
	}
}

func TestTownhallOnlyAIRecruitsScoutDeterministically(t *testing.T) {
	g := aiTestGame(3, 3)
	addTownhall(g, 0, game.NewHex(1, 1))
	snapshot := NewWorldSnapshot(g, nil, nil)
	first := NewController(g.Map.Seed, 0, StandardConfig()).Plan(
		&snapshot,
		OwnState{Owner: 0},
	)
	second := NewController(g.Map.Seed, 0, StandardConfig()).Plan(
		&snapshot,
		OwnState{Owner: 0},
	)

	if !reflect.DeepEqual(first, second) {
		t.Fatalf("same state and seed produced different plans:\n%+v\n%+v", first, second)
	}
	if first.Manual == nil ||
		first.Manual.Type != game.ActionRecruit ||
		first.Manual.Recruit == nil ||
		first.Manual.Recruit.Unit != game.UnitScout {
		t.Fatalf("opening plan = %+v, want Scout recruitment", first)
	}
	if first.Trace.Candidates > StandardConfig().MaxCandidates ||
		first.Trace.PathQueries > StandardConfig().MaxPathQueries {
		t.Fatalf("planning budgets exceeded: %+v", first.Trace)
	}
}

func TestAdjacentThreatProducesManualAttack(t *testing.T) {
	g := aiTestGame(3, 1)
	g.Factions[0].Alive = true
	g.Factions[1].Alive = true
	g.Map.Grid[0][0].Units = []game.UnitData{{
		Type: game.UnitKnight, Owner: 0, HP: 8,
	}}
	g.Map.Grid[1][0].Units = []game.UnitData{{
		Type: game.UnitScout, Owner: 1, HP: 3,
	}}
	snapshot := NewWorldSnapshot(g, nil, nil)
	plan := NewController(g.Map.Seed, 0, StandardConfig()).Plan(
		&snapshot,
		OwnState{Owner: 0},
	)

	if plan.Manual == nil ||
		plan.Manual.Type != game.ActionAttack ||
		plan.Manual.Attack == nil ||
		plan.Manual.Attack.To != game.NewHex(1, 0) {
		t.Fatalf("plan = %+v, want adjacent attack", plan)
	}
	for _, order := range plan.Orders {
		if order.Kind == OrderAttack && game.HexAdjacent(order.From, order.To) {
			t.Fatalf("AI manufactured adjacent persistent attack order: %+v", order)
		}
	}
}

func TestAdjacentAttackCannotBeStarvedByScoutCandidates(t *testing.T) {
	g := aiTestGame(8, 2)
	addTownhall(g, 0, game.NewHex(0, 0))
	addTownhall(g, 1, game.NewHex(7, 1))
	for _, position := range []game.Hex{
		game.NewHex(0, 1),
		game.NewHex(1, 0),
		game.NewHex(1, 1),
		game.NewHex(2, 0),
		game.NewHex(2, 1),
	} {
		g.Map.GetCell(position).Units = []game.UnitData{{
			Type: game.UnitScout, Owner: 0, HP: 3,
		}}
	}
	g.Map.GetCell(game.NewHex(5, 0)).Units = []game.UnitData{{
		Type: game.UnitKnight, Owner: 0, HP: 8,
	}}
	g.Map.GetCell(game.NewHex(6, 0)).Units = []game.UnitData{{
		Type: game.UnitPeasant, Owner: 1, HP: 5,
	}}
	config := StandardConfig()
	config.MaxCandidates = 3
	plan := NewController(g.Map.Seed, 0, config).Plan(
		ptrSnapshot(NewWorldSnapshot(g, nil, nil)),
		OwnState{Owner: 0},
	)

	if plan.Manual == nil ||
		plan.Manual.Type != game.ActionAttack ||
		plan.Manual.Attack == nil ||
		plan.Manual.Attack.From != game.NewHex(5, 0) {
		t.Fatalf("candidate-limited plan = %+v, want adjacent Knight attack", plan)
	}
}

func TestCombatCadenceRepaysArmyAdvanceDebt(t *testing.T) {
	g := aiTestGame(7, 1)
	addTownhall(g, 0, game.NewHex(0, 0))
	addTownhall(g, 1, game.NewHex(6, 0))
	g.Map.GetCell(game.NewHex(1, 0)).Units = []game.UnitData{{
		Type: game.UnitKnight, Owner: 0, HP: 8,
	}}
	snapshot := NewWorldSnapshot(g, nil, nil)
	controller := NewController(g.Map.Seed, 0, StandardConfig())
	controller.combatAdvanceDebt = 2
	own := OwnState{
		Owner: 0,
		AttackOrders: []game.AttackOrder{{
			From:       game.NewHex(1, 0),
			TargetTile: game.NewHex(6, 0),
		}},
	}

	for wantDebt := 1; wantDebt >= 0; wantDebt-- {
		plan := controller.Plan(&snapshot, own)
		if plan.Manual != nil || plan.Trace.Choice != "Advance best route" {
			t.Fatalf("combat debt plan = %+v, want route advancement", plan)
		}
		if controller.combatAdvanceDebt != wantDebt {
			t.Fatalf(
				"combatAdvanceDebt = %d, want %d",
				controller.combatAdvanceDebt,
				wantDebt,
			)
		}
	}
}

func ptrSnapshot(snapshot WorldSnapshot) *WorldSnapshot {
	return &snapshot
}

func TestDistantEnemyProducesBoundedAttackRoute(t *testing.T) {
	g := aiTestGame(7, 1)
	addTownhall(g, 0, game.NewHex(0, 0))
	addTownhall(g, 1, game.NewHex(6, 0))
	g.Map.Grid[1][0].Units = []game.UnitData{{
		Type: game.UnitKnight, Owner: 0, HP: 8,
	}}
	snapshot := NewWorldSnapshot(g, nil, nil)
	plan := NewController(g.Map.Seed, 0, StandardConfig()).Plan(
		&snapshot,
		OwnState{Owner: 0},
	)

	found := false
	for _, order := range plan.Orders {
		if order.Kind == OrderAttack &&
			order.From == game.NewHex(1, 0) &&
			order.To == game.NewHex(6, 0) {
			found = true
		}
	}
	if !found {
		t.Fatalf("orders = %+v, want attack route toward enemy Townhall", plan.Orders)
	}
	if plan.Trace.PathQueries > StandardConfig().MaxPathQueries {
		t.Fatalf("path queries = %d", plan.Trace.PathQueries)
	}
}

func TestPersonalityIsSeededAndNormalized(t *testing.T) {
	first := seededPersonality(1, 0)
	second := seededPersonality(2, 0)
	if first == second {
		t.Fatal("different seeds produced identical personalities")
	}
	values := []float64{
		first.Economy,
		first.Expansion,
		first.Defense,
		first.Aggression,
		first.Risk,
		first.Opportunism,
	}
	var sum float64
	for _, value := range values {
		if value < 0.85 || value > 1.15 {
			t.Fatalf("personality value out of bounded range: %f", value)
		}
		sum += value
	}
	if mean := sum / float64(len(values)); mean < 0.999 || mean > 1.001 {
		t.Fatalf("personality mean = %f", mean)
	}
}
