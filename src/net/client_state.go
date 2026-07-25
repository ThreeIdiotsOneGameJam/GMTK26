package net

import (
	"sync"

	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/net/packets"
)

type localGameState struct {
	mu         sync.RWMutex
	InGame     bool
	FactionIdx int
	Round      int32
	Deadline   int64
	Map        game.Map
	Coins      int32
	Points     int32
	Resources  game.Resources
}

var LocalGameState = &localGameState{}

func (s *localGameState) ApplyStartPacket(p *packets.S2CGameStartPacket) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.InGame = true
	s.FactionIdx = p.FactionIdx
	s.Round = p.Round
	s.Deadline = p.Deadline
	s.Map = p.Map
	s.Coins = p.Coins
	s.Points = p.Points
	s.Resources = p.Resources
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
}

func (s *localGameState) ApplyEndPacket() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.InGame = false
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

func SendBuildAction(round int32, hex game.Hex, building game.BuildingType) error {
	return Send(&packets.C2SActionPacket{
		Round: round,
		Type:  game.ActionBuild,
		Build: &game.BuildActionPayload{Hex: hex, Building: building},
	})
}

func SendDispatchAction(round int32, hex, to game.Hex, troop game.TroopType) error {
	return Send(&packets.C2SActionPacket{
		Round:    round,
		Type:     game.ActionDispatch,
		Dispatch: &game.DispatchActionPayload{Hex: hex, To: to, Troop: troop},
	})
}
