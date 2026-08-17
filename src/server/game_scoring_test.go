package server

import (
	"reflect"
	"strings"
	"testing"

	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/net/packets"
)

func TestControlScoreSchedule(t *testing.T) {
	for _, round := range []int32{1, 5, 7, 11} {
		if controlScoreDue(round) {
			t.Errorf("controlScoreDue(%d) = true", round)
		}
	}
	for _, round := range []int32{6, 12, 18} {
		if !controlScoreDue(round) {
			t.Errorf("controlScoreDue(%d) = false", round)
		}
	}
}

func TestAwardControlScoreSkipsEliminatedFactions(t *testing.T) {
	g := actionTestGame(2, 1)
	g.Map.Grid[0][0] = game.Cell{
		Tile:     game.TilePlains,
		Owner:    0,
		Building: &game.BuildingData{Type: game.BuildingFarm},
	}
	g.Map.Grid[1][0] = game.Cell{
		Tile:     game.TileGold,
		Owner:    1,
		Building: &game.BuildingData{Type: game.BuildingMine},
	}
	g.Factions[1].Alive = false
	gi := NewGameInstance(1, g, nil)

	gi.awardControlScore()

	if got := g.Factions[0].Points; got != 2 {
		t.Fatalf("living faction Points = %d, want 2", got)
	}
	if got := g.Factions[1].Points; got != 0 {
		t.Fatalf("eliminated faction Points = %d, want 0", got)
	}
}

func TestUnitDestructionAwardsPointsOnce(t *testing.T) {
	g := actionTestGame(2, 1)
	from, to := game.NewHex(0, 0), game.NewHex(1, 0)
	putUnit(g, from, game.UnitPeasant, 0)
	putUnit(g, to, game.UnitScout, 1)
	g.Map.GetCell(to).Units[0].HP = 1
	gi := NewGameInstance(1, g, nil)

	if !gi.resolveAttack(0, from, to) {
		t.Fatal("lethal unit attack did not resolve")
	}
	if got := g.Factions[0].Points; got != game.UnitDestructionScore(game.UnitScout) {
		t.Fatalf("Points = %d, want Scout destruction score", got)
	}
	if gi.resolveAttack(0, from, to) {
		t.Fatal("empty target resolved a second attack")
	}
	if got := g.Factions[0].Points; got != game.UnitDestructionScore(game.UnitScout) {
		t.Fatalf("second attack changed Points to %d", got)
	}
}

func TestBuildingDestructionAwardsPointsOnce(t *testing.T) {
	g := actionTestGame(2, 1)
	from, to := game.NewHex(0, 0), game.NewHex(1, 0)
	putUnit(g, from, game.UnitPeasant, 0)
	g.Map.Grid[1][0] = game.Cell{
		Tile:  game.TilePlains,
		Owner: 1,
		Building: &game.BuildingData{
			Type: game.BuildingFarm,
			HP:   1,
		},
	}
	gi := NewGameInstance(1, g, nil)

	if !gi.resolveAttack(0, from, to) {
		t.Fatal("lethal building attack did not resolve")
	}
	if got := g.Factions[0].Points; got != game.BuildingDestructionScore(game.BuildingFarm, game.TilePlains) {
		t.Fatalf("Points = %d, want Farm destruction score", got)
	}
	if gi.resolveAttack(0, from, to) {
		t.Fatal("empty target resolved a second attack")
	}
	if got := g.Factions[0].Points; got != game.BuildingDestructionScore(game.BuildingFarm, game.TilePlains) {
		t.Fatalf("second attack changed Points to %d", got)
	}
}

func TestEliminatedFactionReceivesNoIncome(t *testing.T) {
	g := actionTestGame(1, 1)
	g.Map.Grid[0][0] = game.Cell{
		Tile:     game.TileGold,
		Owner:    0,
		Building: &game.BuildingData{Type: game.BuildingMine},
	}
	g.Factions[0].Alive = false
	beforeCoins := g.Factions[0].Coins
	beforeResources := cloneResources(g.Factions[0].Resources)
	gi := NewGameInstance(1, g, nil)

	gi.processAutoActions()

	if g.Factions[0].Coins != beforeCoins {
		t.Fatalf("eliminated faction Coins = %d, want %d", g.Factions[0].Coins, beforeCoins)
	}
	if !reflect.DeepEqual(g.Factions[0].Resources, beforeResources) {
		t.Fatalf("eliminated faction Resources = %v, want %v", g.Factions[0].Resources, beforeResources)
	}
}

func TestEliminatedFactionCannotSubmitActions(t *testing.T) {
	g := actionTestGame(1, 1)
	g.Factions[0].Alive = false
	client := NewClient(nil)
	gi := NewGameInstance(1, g, []*Client{client})
	client.JoinGame(gi)

	_, err := client.HandlePacket(&packets.C2SActionPacket{
		Round: 1,
		Type:  game.ActionPass,
	})
	if err == nil || !strings.Contains(err.Error(), "faction has been eliminated") {
		t.Fatalf("eliminated action error = %v", err)
	}
}

func TestBarracksDeductsCoinsWoodAndStone(t *testing.T) {
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

	if g.Map.GetCell(to).BuildingType() != game.BuildingBarracks {
		t.Fatal("Barracks was not constructed")
	}
	if got := g.Factions[0].Coins; got != 76 {
		t.Fatalf("Coins = %d, want 76", got)
	}
	if got := g.Factions[0].Resources[game.ResourceWood]; got != 94 {
		t.Fatalf("Wood = %d, want 94", got)
	}
	if got := g.Factions[0].Resources[game.ResourceStone]; got != 96 {
		t.Fatalf("Stone = %d, want 96", got)
	}
}

func TestBarracksRejectsInsufficientStoneWithoutDeductions(t *testing.T) {
	g := actionTestGame(2, 1)
	from, to := game.NewHex(0, 0), game.NewHex(1, 0)
	putUnit(g, from, game.UnitScout, 0)
	g.Factions[0].Coins = 24
	g.Factions[0].Resources = game.Resources{
		game.ResourceWood:  6,
		game.ResourceStone: 3,
	}
	before := cloneResources(g.Factions[0].Resources)
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

	if g.Map.GetCell(to).HasBuilding() {
		t.Fatal("Barracks was constructed without enough Stone")
	}
	if g.Factions[0].Coins != 24 {
		t.Fatalf("rejected Barracks left Coins = %d, want 24", g.Factions[0].Coins)
	}
	if !reflect.DeepEqual(g.Factions[0].Resources, before) {
		t.Fatalf("rejected Barracks left Resources = %v, want %v", g.Factions[0].Resources, before)
	}
}

func TestRecruitmentDeductsConfiguredResourceCosts(t *testing.T) {
	for _, unit := range []game.UnitType{
		game.UnitPeasant,
		game.UnitArcher,
		game.UnitKnight,
	} {
		t.Run(unit.String(), func(t *testing.T) {
			g := actionTestGame(2, 1)
			from, to := game.NewHex(0, 0), game.NewHex(1, 0)
			g.Map.Grid[0][0] = game.Cell{
				Tile:     game.TilePlains,
				Owner:    0,
				Building: &game.BuildingData{Type: game.BuildingBarracks},
			}
			g.Factions[0].Resources = game.UnitResourceCost(unit)
			gi := NewGameInstance(1, g, nil)
			gi.actions[0] = &submittedAction{
				Type: game.ActionRecruit,
				Recruit: &game.RecruitActionPayload{
					From: from,
					To:   to,
					Unit: unit,
				},
			}

			gi.processClientActions()

			if got := g.Map.GetCell(to).FirstUnit(); got == nil || got.Type != unit {
				t.Fatalf("recruited unit = %+v, want %s", got, unit)
			}
			if got := g.Factions[0].Coins; got != 100-game.UnitCost(unit) {
				t.Fatalf("Coins = %d, want %d", got, 100-game.UnitCost(unit))
			}
			for resource := range game.UnitResourceCost(unit) {
				if got := g.Factions[0].Resources[resource]; got != 0 {
					t.Errorf("%s = %d after recruitment, want 0", resource, got)
				}
			}
		})
	}
}

func cloneResources(resources game.Resources) game.Resources {
	cloned := make(game.Resources, len(resources))
	for resource, amount := range resources {
		cloned[resource] = amount
	}
	return cloned
}
