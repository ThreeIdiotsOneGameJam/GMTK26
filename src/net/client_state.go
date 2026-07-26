package net

import (
	"sync"

	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/net/packets"
)

type localGameState struct {
	mu           sync.RWMutex
	InGame       bool
	FactionIdx   int
	Round        int32
	Deadline     int64
	GameEndTime  int64
	Map          game.Map
	Coins        int32
	Points       int32
	Resources    game.Resources
	Orders       []game.MovementOrder
	AttackOrders []game.AttackOrder
	Result       *game.ActionResult
	Movements    []game.MovementEvent
	AttackEvents []game.AttackEvent
}

var LocalGameState = &localGameState{}

func (s *localGameState) ApplyStartPacket(p *packets.S2CGameStartPacket) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.InGame = true
	s.FactionIdx = p.FactionIdx
	s.Round = p.Round
	s.Deadline = p.Deadline
	s.GameEndTime = p.GameEndTime
	s.Map = p.Map
	s.Coins = p.Coins
	s.Points = p.Points
	s.Resources = p.Resources
	s.Orders = append([]game.MovementOrder(nil), p.Orders...)
	s.AttackOrders = append([]game.AttackOrder(nil), p.AttackOrders...)
	s.Result = nil
	s.Movements = nil
	s.AttackEvents = nil
}

func (s *localGameState) ApplyStatePacket(p *packets.S2CGameStatePacket) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Round = p.Round
	s.Deadline = p.Deadline
	s.Map = p.Map
	s.Coins = p.Coins
	s.Points = p.Points
	s.Resources = p.Resources
	s.Orders = append([]game.MovementOrder(nil), p.Orders...)
	s.AttackOrders = append([]game.AttackOrder(nil), p.AttackOrders...)
	s.Result = p.Result
	s.Movements = append([]game.MovementEvent(nil), p.Movements...)
	s.AttackEvents = append([]game.AttackEvent(nil), p.AttackEvents...)
}

func (s *localGameState) ApplyEndPacket() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.InGame = false
}

func (s *localGameState) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.InGame = false
	s.FactionIdx = 0
	s.Round = 0
	s.Deadline = 0
	s.GameEndTime = 0
	s.Map = game.Map{}
	s.Coins = 0
	s.Points = 0
	s.Resources = nil
	s.Orders = nil
	s.AttackOrders = nil
	s.Result = nil
	s.Movements = nil
	s.AttackEvents = nil
}

func (s *localGameState) GetRound() int32 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Round
}

func (s *localGameState) GetCoins() int32 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Coins
}

func (s *localGameState) GetOrders() []game.MovementOrder {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]game.MovementOrder(nil), s.Orders...)
}

func SendBuildAction(round int32, from, to game.Hex, building game.BuildingType) error {
	return Send(&packets.C2SActionPacket{
		Round: round,
		Type:  game.ActionBuild,
		Build: &game.BuildActionPayload{From: from, To: to, Building: building},
	})
}

func SendMoveAction(round int32, from, to game.Hex) error {
	return Send(&packets.C2SActionPacket{
		Round: round,
		Type:  game.ActionMove,
		Move:  &game.MoveActionPayload{From: from, To: to},
	})
}

func SendRecruitAction(round int32, from, to game.Hex, unit game.UnitType) error {
	return Send(&packets.C2SActionPacket{
		Round:   round,
		Type:    game.ActionRecruit,
		Recruit: &game.RecruitActionPayload{From: from, To: to, Unit: unit},
	})
}

func SendAttackAction(round int32, from, to game.Hex) error {
	return Send(&packets.C2SActionPacket{
		Round:  round,
		Type:   game.ActionAttack,
		Attack: &game.AttackActionPayload{From: from, To: to},
	})
}

func SendPassAction(round int32) error {
	return Send(&packets.C2SActionPacket{
		Round: round,
		Type:  game.ActionPass,
	})
}

func SendCancelMovementOrder(round int32, from game.Hex) error {
	return Send(&packets.C2SCancelMovementOrderPacket{
		Round: round,
		From:  from,
	})
}

func SendCancelBuildAction(round int32, to game.Hex) error {
	return Send(&packets.C2SCancelBuildActionPacket{
		Round: round,
		To:    to,
	})
}
