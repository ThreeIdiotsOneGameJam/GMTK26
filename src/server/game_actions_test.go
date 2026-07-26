package server

import (
	"testing"

	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/net/packets"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

func actionTestGame(width, height int32) *game.Game {
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
			Grid:     grid,
			GridSize: vec.Vec2i{X: width, Y: height},
		},
	}
	for i := range g.Factions {
		g.Factions[i].Coins = 100
		g.Factions[i].Alive = true
		g.Factions[i].Resources = game.Resources{
			game.ResourceWood:  100,
			game.ResourceStone: 100,
			game.ResourceIron:  100,
		}
	}
	return g
}

func putUnit(g *game.Game, hex game.Hex, unit game.UnitType, owner int8) {
	cell := g.Map.GetCell(hex)
	stats := game.GetUnitStats(unit)
	cell.Units = []game.UnitData{{Type: unit, Owner: owner, HP: stats.MaxHP}}
}

func TestAssignedRouteAdvancesAndPersists(t *testing.T) {
	g := actionTestGame(1, 5)
	from, destination := game.NewHex(0, 0), game.NewHex(0, 4)
	putUnit(g, from, game.UnitPeasant, 0)
	gi := NewGameInstance(1, g, nil)
	if err := gi.setMovementOrderLocked(0, game.MoveActionPayload{
		From: from,
		To:   destination,
	}); err != nil {
		t.Fatal(err)
	}

	gi.processClientActions()

	endpoint := game.NewHex(0, 2)
	if !g.Map.GetCell(endpoint).HasUnits() ||
		g.Map.GetCell(endpoint).Units[0].Type != game.UnitPeasant ||
		g.Map.GetCell(endpoint).Units[0].Owner != 0 ||
		g.Map.GetCell(from).HasUnits() {
		t.Fatalf("unit did not move to %v", endpoint)
	}
	orders := gi.movementOrders[0]
	if len(orders) != 1 ||
		orders[0].Current != endpoint ||
		orders[0].Destination != destination {
		t.Fatalf("orders = %v", orders)
	}
	if len(gi.movementEvents) != 1 || len(gi.movementEvents[0].Path) != 3 {
		t.Fatalf("movement events = %v", gi.movementEvents)
	}
}

func TestActionPacketReachesAuthoritativeMovementResolver(t *testing.T) {
	g := actionTestGame(1, 4)
	from, destination := game.NewHex(0, 0), game.NewHex(0, 3)
	putUnit(g, from, game.UnitScout, 0)
	client := NewClient(nil)
	gi := NewGameInstance(1, g, []*Client{client})
	client.JoinGame(gi)

	response, err := client.HandlePacket(&packets.C2SActionPacket{
		Round: 1,
		Type:  game.ActionMove,
		Move: &game.MoveActionPayload{
			From: from,
			To:   destination,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if response != nil {
		t.Fatalf("unexpected response %T", response)
	}
	if _, exists := gi.actions[0]; exists {
		t.Fatal("route assignment occupied the faction's manual action slot")
	}

	gi.processClientActions()

	if !g.Map.GetCell(destination).HasUnits() || g.Map.GetCell(destination).Units[0].Type != game.UnitScout {
		t.Fatal("packet-submitted movement did not resolve")
	}
	result := gi.actionResults[0]
	if result == nil || result.Status != game.ActionResultSucceeded {
		t.Fatalf("result = %+v", result)
	}
}

func TestNewRouteAdvancesBeforeExistingRoundRobinQueue(t *testing.T) {
	g := actionTestGame(2, 2)
	oldFrom, oldTo := game.NewHex(0, 0), game.NewHex(0, 1)
	newFrom, newTo := game.NewHex(1, 0), game.NewHex(1, 1)
	putUnit(g, oldFrom, game.UnitScout, 0)
	putUnit(g, newFrom, game.UnitScout, 0)
	gi := NewGameInstance(1, g, nil)
	gi.movementOrders[0] = []game.MovementOrder{{
		Current:     oldFrom,
		Destination: oldTo,
	}}
	if err := gi.setMovementOrderLocked(0, game.MoveActionPayload{
		From: newFrom,
		To:   newTo,
	}); err != nil {
		t.Fatal(err)
	}

	gi.processClientActions()

	if !g.Map.GetCell(newTo).HasUnits() || g.Map.GetCell(newTo).Units[0].Type != game.UnitScout {
		t.Fatal("newly assigned route did not receive first advancement")
	}
	if !g.Map.GetCell(oldFrom).HasUnits() || g.Map.GetCell(oldFrom).Units[0].Type != game.UnitScout {
		t.Fatal("existing round-robin route advanced before the new route")
	}
}

func TestManualActionSuppressesNewRouteWithoutDeletingIt(t *testing.T) {
	g := actionTestGame(2, 3)
	from := game.NewHex(0, 0)
	destination := game.NewHex(0, 2)
	buildTarget := game.NewHex(1, 0)
	putUnit(g, from, game.UnitScout, 0)
	gi := NewGameInstance(1, g, nil)
	if err := gi.setMovementOrderLocked(0, game.MoveActionPayload{
		From: from,
		To:   destination,
	}); err != nil {
		t.Fatal(err)
	}
	gi.actions[0] = &submittedAction{
		Type: game.ActionBuild,
		Build: &game.BuildActionPayload{
			From:     from,
			To:       buildTarget,
			Building: game.BuildingFarm,
		},
	}

	gi.processClientActions()

	if !g.Map.GetCell(from).HasUnits() || g.Map.GetCell(from).Units[0].Type != game.UnitScout {
		t.Fatal("manual build did not suppress route advancement")
	}
	if len(gi.movementOrders[0]) != 1 {
		t.Fatal("manual build deleted the newly assigned route")
	}
	if _, exists := gi.movementPriorities[0]; exists {
		t.Fatal("new-route priority leaked into the next round")
	}

	delete(gi.actions, 0)
	gi.processClientActions()
	if !g.Map.GetCell(destination).HasUnits() || g.Map.GetCell(destination).Units[0].Type != game.UnitScout {
		t.Fatal("persisted route did not advance on the next free round")
	}
}

func TestAutomaticMovementSkipsBlockedOrder(t *testing.T) {
	g := actionTestGame(2, 4)
	blockedFrom := game.NewHex(0, 0)
	blockedDestination := game.NewHex(0, 1)
	movableFrom := game.NewHex(0, 3)
	movableDestination := game.NewHex(1, 3)
	putUnit(g, blockedFrom, game.UnitScout, 0)
	putUnit(g, blockedDestination, game.UnitPeasant, 0)
	putUnit(g, movableFrom, game.UnitScout, 0)
	gi := NewGameInstance(1, g, nil)
	gi.movementOrders[0] = []game.MovementOrder{
		{Current: blockedFrom, Destination: blockedDestination},
		{Current: movableFrom, Destination: movableDestination},
	}

	gi.processClientActions()

	if !g.Map.GetCell(movableDestination).HasUnits() || g.Map.GetCell(movableDestination).Units[0].Type != game.UnitScout {
		t.Fatal("movable order did not advance")
	}
	if got := gi.movementOrders[0]; len(got) != 1 || got[0].Current != blockedFrom {
		t.Fatalf("remaining orders = %v", got)
	}
}

func TestPassSuppressesAutomaticMovement(t *testing.T) {
	g := actionTestGame(1, 3)
	from, destination := game.NewHex(0, 0), game.NewHex(0, 2)
	putUnit(g, from, game.UnitScout, 0)
	gi := NewGameInstance(1, g, nil)
	gi.movementOrders[0] = []game.MovementOrder{{
		Current:     from,
		Destination: destination,
	}}
	gi.actions[0] = &submittedAction{Type: game.ActionPass}

	gi.processClientActions()

	if !g.Map.GetCell(from).HasUnits() || g.Map.GetCell(from).Units[0].Type != game.UnitScout {
		t.Fatal("pass allowed automatic movement")
	}
	if len(gi.movementOrders[0]) != 1 {
		t.Fatal("pass removed movement order")
	}
}

func TestAllBlockedOrdersReportBlocked(t *testing.T) {
	g := actionTestGame(1, 2)
	from, destination := game.NewHex(0, 0), game.NewHex(0, 1)
	putUnit(g, from, game.UnitScout, 0)
	putUnit(g, destination, game.UnitPeasant, 0)
	gi := NewGameInstance(1, g, nil)
	gi.movementOrders[0] = []game.MovementOrder{{
		Current:     from,
		Destination: destination,
	}}

	gi.processClientActions()

	result := gi.actionResults[0]
	if result == nil || result.Status != game.ActionResultBlocked || !result.Automatic {
		t.Fatalf("result = %+v", result)
	}
	if len(gi.movementOrders[0]) != 1 {
		t.Fatal("blocked order was removed")
	}
}

func TestFreeCancellationWithdrawsNewRoutePriority(t *testing.T) {
	g := actionTestGame(1, 3)
	from := game.NewHex(0, 0)
	putUnit(g, from, game.UnitScout, 0)
	client := NewClient(nil)
	gi := NewGameInstance(1, g, []*Client{client})
	client.JoinGame(gi)
	if err := gi.setMovementOrderLocked(0, game.MoveActionPayload{
		From: from,
		To:   game.NewHex(0, 2),
	}); err != nil {
		t.Fatal(err)
	}

	if err := gi.CancelMovementOrder(client, 1, from); err != nil {
		t.Fatal(err)
	}
	if len(gi.movementOrders[0]) != 0 {
		t.Fatal("movement order was not cancelled")
	}
	if _, exists := gi.movementPriorities[0]; exists {
		t.Fatal("new route priority was not withdrawn")
	}
}

func TestCancelPendingBuildRestoresAutomaticMovement(t *testing.T) {
	g := actionTestGame(2, 2)
	moveFrom := game.NewHex(0, 0)
	moveTo := game.NewHex(0, 1)
	buildFrom := game.NewHex(1, 0)
	buildTo := game.NewHex(1, 1)
	putUnit(g, moveFrom, game.UnitPeasant, 0)
	putUnit(g, buildFrom, game.UnitScout, 0)

	client := NewClient(nil)
	gi := NewGameInstance(1, g, []*Client{client})
	client.JoinGame(gi)
	gi.movementOrders[0] = []game.MovementOrder{{
		Current:     moveFrom,
		Destination: moveTo,
	}}
	gi.actions[0] = &submittedAction{
		Type: game.ActionBuild,
		Build: &game.BuildActionPayload{
			From:     buildFrom,
			To:       buildTo,
			Building: game.BuildingFarm,
		},
	}

	response, err := client.HandlePacket(&packets.C2SCancelBuildActionPacket{
		Round: 1,
		To:    buildTo,
	})
	if err != nil {
		t.Fatal(err)
	}
	if response != nil {
		t.Fatalf("unexpected response %T", response)
	}
	if _, exists := gi.actions[0]; exists {
		t.Fatal("pending build action was not withdrawn")
	}

	gi.processClientActions()

	if g.Map.GetCell(buildTo).HasBuilding() {
		t.Fatal("cancelled building was constructed")
	}
	if !g.Map.GetCell(moveTo).HasUnits() || g.Map.GetCell(moveTo).Units[0].Type != game.UnitPeasant {
		t.Fatal("cancelling the build did not restore automatic movement")
	}
}

func TestStaleBuildCancelDoesNotEraseReplacementAction(t *testing.T) {
	g := actionTestGame(2, 2)
	client := NewClient(nil)
	gi := NewGameInstance(1, g, []*Client{client})
	replacementTarget := game.NewHex(1, 1)
	gi.actions[0] = &submittedAction{
		Type: game.ActionBuild,
		Build: &game.BuildActionPayload{
			From:     game.NewHex(1, 0),
			To:       replacementTarget,
			Building: game.BuildingFarm,
		},
	}

	if err := gi.CancelBuildAction(client, 1, game.NewHex(0, 1)); err != nil {
		t.Fatal(err)
	}
	if _, exists := gi.actions[0]; !exists {
		t.Fatal("stale target erased a replacement build")
	}
}

func TestCancelBuildAdvancesRouteAssignedInSameRound(t *testing.T) {
	g := actionTestGame(2, 3)
	from := game.NewHex(0, 0)
	destination := game.NewHex(0, 2)
	buildTarget := game.NewHex(1, 0)
	putUnit(g, from, game.UnitScout, 0)

	client := NewClient(nil)
	gi := NewGameInstance(1, g, []*Client{client})
	client.JoinGame(gi)

	if _, err := client.HandlePacket(&packets.C2SActionPacket{
		Round: 1,
		Type:  game.ActionMove,
		Move: &game.MoveActionPayload{
			From: from,
			To:   destination,
		},
	}); err != nil {
		t.Fatal(err)
	}
	if len(gi.movementOrders[0]) != 1 {
		t.Fatal("route assignment did not persist immediately")
	}

	if _, err := client.HandlePacket(&packets.C2SActionPacket{
		Round: 1,
		Type:  game.ActionBuild,
		Build: &game.BuildActionPayload{
			From:     from,
			To:       buildTarget,
			Building: game.BuildingFarm,
		},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.HandlePacket(&packets.C2SCancelBuildActionPacket{
		Round: 1,
		To:    buildTarget,
	}); err != nil {
		t.Fatal(err)
	}

	gi.processClientActions()

	if !g.Map.GetCell(destination).HasUnits() || g.Map.GetCell(destination).Units[0].Type != game.UnitScout {
		t.Fatal("same-round route did not advance after pending build cancellation")
	}
	if g.Map.GetCell(buildTarget).HasBuilding() {
		t.Fatal("cancelled build resolved")
	}
}

func TestContestedMovementCancelsAllOrders(t *testing.T) {
	g := actionTestGame(3, 1)
	left, target, right := game.NewHex(0, 0), game.NewHex(1, 0), game.NewHex(2, 0)
	putUnit(g, left, game.UnitScout, 0)
	putUnit(g, right, game.UnitScout, 1)
	gi := NewGameInstance(1, g, nil)
	gi.movementOrders[0] = []game.MovementOrder{{Current: left, Destination: target}}
	gi.movementOrders[1] = []game.MovementOrder{{Current: right, Destination: target}}

	gi.processClientActions()

	if !g.Map.GetCell(left).HasUnits() ||
		!g.Map.GetCell(right).HasUnits() ||
		g.Map.GetCell(target).HasUnits() {
		t.Fatal("contested movement mutated unit positions")
	}
	if len(gi.movementOrders[0]) != 0 || len(gi.movementOrders[1]) != 0 {
		t.Fatal("contested movement orders were not cancelled")
	}
	if gi.actionResults[0].Status != game.ActionResultContested ||
		gi.actionResults[1].Status != game.ActionResultContested {
		t.Fatalf("unexpected results: %v %v", gi.actionResults[0], gi.actionResults[1])
	}
}

func TestMixedBuildAndMoveConflictFailsBoth(t *testing.T) {
	g := actionTestGame(3, 2)
	moveFrom := game.NewHex(0, 0)
	target := game.NewHex(1, 0)
	buildFrom := game.NewHex(2, 0)
	putUnit(g, moveFrom, game.UnitPeasant, 0)
	putUnit(g, buildFrom, game.UnitScout, 1)
	gi := NewGameInstance(1, g, nil)
	gi.movementOrders[0] = []game.MovementOrder{{Current: moveFrom, Destination: target}}
	gi.actions[1] = &submittedAction{
		Type: game.ActionBuild,
		Build: &game.BuildActionPayload{
			From:     buildFrom,
			To:       target,
			Building: game.BuildingBarracks,
		},
	}

	gi.processClientActions()

	if g.Map.GetCell(moveFrom).HasUnits() == false ||
		g.Map.GetCell(target).HasUnits() ||
		g.Map.GetCell(target).HasBuilding() {
		t.Fatal("mixed conflict mutated target")
	}
	if g.Factions[1].Coins != 100 {
		t.Fatal("contested build deducted coins")
	}
	if g.Factions[1].Resources[game.ResourceWood] != 100 ||
		g.Factions[1].Resources[game.ResourceStone] != 100 {
		t.Fatal("contested build deducted resources")
	}
	if gi.actionResults[0].Status != game.ActionResultContested ||
		gi.actionResults[1].Status != game.ActionResultContested {
		t.Fatalf("unexpected results: %v %v", gi.actionResults[0], gi.actionResults[1])
	}
}

func TestRecruitDoesNotClaimDestination(t *testing.T) {
	g := actionTestGame(2, 1)
	from, to := game.NewHex(0, 0), game.NewHex(1, 0)
	source := g.Map.GetCell(from)
	source.Owner = 0
	source.Building = &game.BuildingData{Type: game.BuildingTownhall, HP: game.BuildingMaxHP(game.BuildingTownhall)}
	gi := NewGameInstance(1, g, nil)
	gi.actions[0] = &submittedAction{
		Type: game.ActionRecruit,
		Recruit: &game.RecruitActionPayload{
			From: from,
			To:   to,
			Unit: game.UnitScout,
		},
	}

	gi.processClientActions()

	target := g.Map.GetCell(to)
	if !target.HasUnits() || target.Units[0].Type != game.UnitScout || target.Units[0].Owner != 0 || target.Owner != -1 {
		t.Fatalf("recruitment target = %+v", target)
	}
	if g.Factions[0].Coins != 92 {
		t.Fatalf("coins = %d, want 92", g.Factions[0].Coins)
	}
}

func TestRecruitUsesCurrentRoundFarmFood(t *testing.T) {
	g := actionTestGame(3, 1)
	from, to, farm := game.NewHex(0, 0), game.NewHex(1, 0), game.NewHex(2, 0)
	source := g.Map.GetCell(from)
	source.Owner = 0
	source.Building = &game.BuildingData{Type: game.BuildingBarracks, HP: game.BuildingMaxHP(game.BuildingBarracks)}
	farmCell := g.Map.GetCell(farm)
	farmCell.Owner = 0
	farmCell.Building = &game.BuildingData{Type: game.BuildingFarm, HP: game.BuildingMaxHP(game.BuildingFarm)}
	g.Factions[0].Resources[game.ResourceFood] = 3

	gi := NewGameInstance(1, g, nil)
	gi.actions[0] = &submittedAction{
		Type: game.ActionRecruit,
		Recruit: &game.RecruitActionPayload{
			From: from,
			To:   to,
			Unit: game.UnitPeasant,
		},
	}

	gi.processAutoActions()
	gi.processClientActions()

	if g.Map.GetCell(to).HasUnits() && g.Map.GetCell(to).Units[0].Type != game.UnitPeasant {
		t.Fatal("incoming Farm food did not fund recruitment")
	}
	if got := g.Factions[0].Resources[game.ResourceFood]; got != 0 {
		t.Fatalf("Food = %d, want 0 after producing 1 and spending 4", got)
	}
	if g.Factions[0].Coins != 92 {
		t.Fatalf("coins = %d, want 92", g.Factions[0].Coins)
	}
}

func TestRecruitRejectsInsufficientFood(t *testing.T) {
	g := actionTestGame(2, 1)
	from, to := game.NewHex(0, 0), game.NewHex(1, 0)
	source := g.Map.GetCell(from)
	source.Owner = 0
	source.Building = &game.BuildingData{Type: game.BuildingBarracks, HP: game.BuildingMaxHP(game.BuildingBarracks)}
	g.Factions[0].Resources = game.Resources{game.ResourceFood: 2}

	gi := NewGameInstance(1, g, nil)
	gi.actions[0] = &submittedAction{
		Type: game.ActionRecruit,
		Recruit: &game.RecruitActionPayload{
			From: from,
			To:   to,
			Unit: game.UnitArcher,
		},
	}

	gi.processClientActions()

	if g.Map.GetCell(to).HasUnits() {
		t.Fatal("Archer was recruited without enough Food")
	}
	if g.Factions[0].Resources[game.ResourceFood] != 2 || g.Factions[0].Coins != 100 {
		t.Fatal("rejected recruitment deducted resources")
	}
	result := gi.actionResults[0]
	if result == nil ||
		result.Status != game.ActionResultInvalid ||
		result.Message != "Not enough Food" {
		t.Fatalf("result = %+v, want insufficient Food", result)
	}
}

func TestScoutBuildsAdjacentAndClaims(t *testing.T) {
	g := actionTestGame(2, 1)
	from, to := game.NewHex(0, 0), game.NewHex(1, 0)
	putUnit(g, from, game.UnitScout, 0)
	gi := NewGameInstance(1, g, nil)
	gi.actions[0] = &submittedAction{
		Type: game.ActionBuild,
		Build: &game.BuildActionPayload{
			From:     from,
			To:       to,
			Building: game.BuildingBarracks,
		},
	}

	gi.processClientActions()

	target := g.Map.GetCell(to)
	if target.BuildingType() != game.BuildingBarracks || target.Owner != 0 {
		t.Fatalf("build target = %+v", target)
	}
	if !g.Map.GetCell(from).HasUnits() || g.Map.GetCell(from).Units[0].Type != game.UnitScout {
		t.Fatal("Scout was consumed by construction")
	}
}

func TestAttackDamagesUnitOverBuilding(t *testing.T) {
	g := actionTestGame(2, 1)
	from, to := game.NewHex(0, 0), game.NewHex(1, 0)
	putUnit(g, from, game.UnitPeasant, 0)
	target := g.Map.GetCell(to)
	target.Owner = 1
	target.Building = &game.BuildingData{Type: game.BuildingBarracks, HP: game.BuildingMaxHP(game.BuildingBarracks)}
	target.Units = []game.UnitData{{Type: game.UnitKnight, Owner: 1, HP: game.GetUnitStats(game.UnitKnight).MaxHP}}
	gi := NewGameInstance(1, g, nil)
	gi.actions[0] = &submittedAction{
		Type:   game.ActionAttack,
		Attack: &game.AttackActionPayload{From: from, To: to},
	}

	gi.processClientActions()

	if !target.HasBuilding() || target.BuildingType() != game.BuildingBarracks {
		t.Fatal("attack hit building instead of enemy unit")
	}
	if target.Owner != 1 {
		t.Fatal("territory ownership changed when building was not destroyed")
	}
	if !target.HasUnits() || target.Units[0].Type != game.UnitKnight || target.Units[0].Owner != 1 {
		t.Fatal("enemy unit was destroyed")
	}
	wantHP := game.GetUnitStats(game.UnitKnight).MaxHP - game.GetUnitStats(game.UnitPeasant).Attack
	if target.Units[0].HP != wantHP {
		t.Fatalf("enemy unit HP = %d, want %d", target.Units[0].HP, wantHP)
	}
	if gi.actionResults[0].Status != game.ActionResultSucceeded {
		t.Fatalf("result = %+v", gi.actionResults[0])
	}
}

func TestTownhallTakesDamage(t *testing.T) {
	g := actionTestGame(2, 1)
	from, to := game.NewHex(0, 0), game.NewHex(1, 0)
	putUnit(g, from, game.UnitKnight, 0)
	target := g.Map.GetCell(to)
	target.Owner = 1
	target.Building = &game.BuildingData{Type: game.BuildingTownhall, HP: game.BuildingMaxHP(game.BuildingTownhall)}
	gi := NewGameInstance(1, g, nil)
	gi.actions[0] = &submittedAction{
		Type:   game.ActionAttack,
		Attack: &game.AttackActionPayload{From: from, To: to},
	}

	gi.processClientActions()

	if !target.HasBuilding() || target.BuildingType() != game.BuildingTownhall {
		t.Fatal("Townhall was demolished in one hit")
	}
	if target.Building.HP <= 0 {
		t.Fatal("Townhall HP should be positive")
	}
	if target.Building.HP >= game.BuildingMaxHP(game.BuildingTownhall) {
		t.Fatal("Townhall took no damage")
	}
	if gi.actionResults[0].Status != game.ActionResultSucceeded {
		t.Fatalf("result = %+v", gi.actionResults[0])
	}
}

func TestScoutCannotAttackBuilding(t *testing.T) {
	g := actionTestGame(2, 1)
	from, to := game.NewHex(0, 0), game.NewHex(1, 0)
	putUnit(g, from, game.UnitScout, 0)
	target := g.Map.GetCell(to)
	target.Owner = 1
	target.Building = &game.BuildingData{Type: game.BuildingBarracks, HP: game.BuildingMaxHP(game.BuildingBarracks)}
	gi := NewGameInstance(1, g, nil)
	gi.actions[0] = &submittedAction{
		Type:   game.ActionAttack,
		Attack: &game.AttackActionPayload{From: from, To: to},
	}

	gi.processClientActions()

	if !target.HasBuilding() || target.BuildingType() != game.BuildingBarracks {
		t.Fatal("Scout demolished building")
	}
	if gi.actionResults[0].Status != game.ActionResultInvalid {
		t.Fatalf("result = %+v", gi.actionResults[0])
	}
}
