package server

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/net/packets"
)

type submittedAction struct {
	Type     game.ActionType
	Build    *game.BuildActionPayload
	Dispatch *game.DispatchActionPayload
}

type GameInstance struct {
	ID             uint64
	game           *game.Game
	clients        []*Client
	factionClients map[*Client]int
	actions        map[int]*submittedAction
	clientsChanged chan struct{}
	done           chan struct{}
	mu             sync.RWMutex
}

func NewGameInstance(id uint64, g *game.Game, clients []*Client) *GameInstance {
	factionClients := make(map[*Client]int)
	for i, c := range clients {
		if c != nil {
			factionClients[c] = i
		}
	}

	return &GameInstance{
		ID:             id,
		game:           g,
		clients:        clients,
		factionClients: factionClients,
		actions:        make(map[int]*submittedAction),
		clientsChanged: make(chan struct{}, 1),
		done:           make(chan struct{}),
	}
}

func (gi *GameInstance) SubmitAction(c *Client, round int32, actionType game.ActionType, build *game.BuildActionPayload, dispatch *game.DispatchActionPayload) error {
	gi.mu.Lock()
	defer gi.mu.Unlock()

	if round != gi.game.Round {
		return fmt.Errorf("wrong round: got %d, want %d", round, gi.game.Round)
	}

	factionIdx, ok := gi.factionClients[c]
	if !ok {
		return fmt.Errorf("client not in this game")
	}

	gi.actions[factionIdx] = &submittedAction{
		Type:     actionType,
		Build:    build,
		Dispatch: dispatch,
	}
	return nil
}

func (gi *GameInstance) sendToClient(client *Client, packet packets.S2CPacket) {
	if client == nil || client.GameInstance() != gi {
		return
	}
	if err := client.SendPacket(packet); err != nil {
		log.Printf("game %d: failed to send %T: %v", gi.ID, packet, err)
	}
}

func (gi *GameInstance) Run() {
	defer func() {
		for _, c := range gi.clients {
			if c != nil {
				c.LeaveGameInstance(gi)
			}
		}
		GameInstances.RemoveGame(gi.ID)
		close(gi.done)
	}()

	gi.game.Map.Generate()

	for i := range gi.game.Factions {
		if i < len(gi.clients) {
			f := &gi.game.Factions[i]
			f.Index = i
			f.Coins = 100
			f.Points = 0
			f.Resources = make(game.Resources)
			f.Alive = true
			if gi.clients[i] == nil {
				f.AI = true
			}
		}
	}

	gi.assignStartingCells()

	gameEndTime := time.Now().Add(5 * time.Minute)
	firstDeadline := time.Now().Add(5 * time.Second)

	gi.setRound(1)

	for i, c := range gi.clients {
		if c == nil {
			continue
		}
		f := gi.game.Factions[i]
		startPacket := &packets.S2CGameStartPacket{
			FactionIdx: i,
			Map:        gi.game.Map,
			Coins:      f.Coins,
			Points:     f.Points,
			Resources:  f.Resources,
			Round:      1,
			Deadline:   firstDeadline.UnixNano(),
		}
		gi.sendToClient(c, startPacket)
	}

	for {
		if !gi.hasConnectedPlayers() {
			return
		}

		roundStart := time.Now()
		deadline := roundStart.Add(5 * time.Second)

		for i, c := range gi.clients {
			if c == nil {
				continue
			}
			f := gi.game.Factions[i]
			statePacket := &packets.S2CGameStatePacket{
				Round:     gi.game.Round,
				Deadline:  deadline.UnixNano(),
				Map:       gi.game.Map,
				Coins:     f.Coins,
				Points:    f.Points,
				Resources: f.Resources,
			}
			gi.sendToClient(c, statePacket)
		}

		if !gi.waitUntil(deadline) {
			return
		}

		gi.processAutoActions()
		gi.processClientActions()

		gi.mu.Lock()
		gi.game.Round++
		gi.actions = make(map[int]*submittedAction)
		gi.mu.Unlock()

		aliveCount := gi.checkAlive()

		if time.Now().After(gameEndTime) || aliveCount <= 1 {
			gi.broadcastGameEnd()
			return
		}
	}
}

// setRound updates the round under the same lock SubmitAction uses to
// validate incoming actions.
func (gi *GameInstance) setRound(round int32) {
	gi.mu.Lock()
	gi.game.Round = round
	gi.mu.Unlock()
}

// assignStartingCells claims one spaced-out land cell per player faction so
// everyone starts with territory; without it checkAlive would eliminate all
// factions after the first round.
func (gi *GameInstance) assignStartingCells() {
	m := &gi.game.Map

	candidates := make([]game.Hex, 0)
	for x := range m.Grid {
		for y := range m.Grid[x] {
			tile := m.Grid[x][y].Tile
			if tile != game.TileVoid && tile != game.TileWater {
				candidates = append(candidates, game.NewHex(int32(x), int32(y)))
			}
		}
	}
	if len(candidates) == 0 {
		return
	}

	r := rand.New(rand.NewSource(m.Seed))
	r.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	spacing := float64(min(m.GridSize.X, m.GridSize.Y)) / 4.0
	claimed := make([]game.Hex, 0, len(gi.clients))

	for i := range gi.game.Factions {
		if i >= len(gi.clients) {
			continue
		}
		hex := pickStartingHex(candidates, claimed, spacing)
		claimed = append(claimed, hex)
		cell := m.GetCell(hex)
		cell.Owner = int8(i)
		cell.Building = game.BuildingTownhall
	}
}

// pickStartingHex returns the first candidate at least spacing away from all
// claimed hexes, halving the requirement until one qualifies.
func pickStartingHex(candidates, claimed []game.Hex, spacing float64) game.Hex {
	for spacing >= 1 {
		for _, hex := range candidates {
			farEnough := true
			for _, other := range claimed {
				dx := float64(hex.X - other.X)
				dy := float64(hex.Y - other.Y)
				if math.Hypot(dx, dy) < spacing {
					farEnough = false
					break
				}
			}
			if farEnough {
				return hex
			}
		}
		spacing /= 2
	}
	return candidates[0]
}

func (gi *GameInstance) hasConnectedPlayers() bool {
	for _, c := range gi.clients {
		if c != nil && c.GameInstance() == gi {
			return true
		}
	}
	return false
}

// clientLeft wakes the game loop so the final client departure terminates the
// instance immediately instead of waiting for the current round deadline.
func (gi *GameInstance) clientLeft() {
	select {
	case gi.clientsChanged <- struct{}{}:
	default:
	}
}

func (gi *GameInstance) waitUntil(deadline time.Time) bool {
	delay := time.Until(deadline)
	if delay <= 0 {
		return gi.hasConnectedPlayers()
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			return gi.hasConnectedPlayers()
		case <-gi.clientsChanged:
			if !gi.hasConnectedPlayers() {
				return false
			}
		}
	}
}

func (gi *GameInstance) processAutoActions() {
	for i := range gi.game.Factions {
		for x := range gi.game.Map.Grid {
			for y := range gi.game.Map.Grid[x] {
				cell := &gi.game.Map.Grid[x][y]
				if cell.Owner == int8(i) && cell.Building != game.BuildingUnknown {
					produced := game.BuildingProduces(cell.Building, cell.Tile)
					for resType, amount := range produced {
						gi.game.Factions[i].Resources[resType] += amount
					}
					gi.game.Factions[i].Coins += game.BuildingCoinsProduces(cell.Building)
				}
			}
		}
	}
}

func (gi *GameInstance) processClientActions() {
	for i := range gi.game.Factions {
		gi.mu.RLock()
		act, submitted := gi.actions[i]
		gi.mu.RUnlock()

		if !submitted || act == nil || act.Type == game.ActionPass {
			continue
		}

		faction := &gi.game.Factions[i]

		switch act.Type {
		case game.ActionBuild:
			payload := act.Build
			if payload == nil {
				continue
			}

			cell := gi.game.Map.GetCell(payload.Hex)
			if cell == nil {
				continue
			}
			if !game.BuildingCanPlace(&gi.game.Map, payload.Building, payload.Hex) {
				continue
			}
			if cell.Owner != -1 && cell.Owner != int8(i) {
				continue
			}
			cost := game.BuildingCost(payload.Building)
			if faction.Coins < cost {
				continue
			}

			faction.Coins -= cost
			cell.Owner = int8(i)
			cell.Building = payload.Building

		case game.ActionDispatch:
			payload := act.Dispatch
			if payload == nil {
				continue
			}

			srcCell := gi.game.Map.GetCell(payload.Hex)
			dstCell := gi.game.Map.GetCell(payload.To)
			if srcCell == nil || dstCell == nil {
				continue
			}
			if srcCell.Owner != int8(i) {
				continue
			}

			if srcCell.Troop == game.TroopUnknown {
				if srcCell.Building != game.BuildingBarracks {
					continue
				}
				cost := game.TroopCost(payload.Troop)
				if faction.Coins < cost {
					continue
				}
				if dstCell.Troop != game.TroopUnknown {
					continue
				}
				faction.Coins -= cost
				dstCell.Troop = payload.Troop
				dstCell.Owner = int8(i)
				continue
			}

			if dstCell.Troop != game.TroopUnknown {
				continue
			}

			dstCell.Troop = srcCell.Troop
			dstCell.Owner = int8(i)
			srcCell.Troop = game.TroopUnknown
		}
	}
}

func (gi *GameInstance) checkAlive() int {
	aliveCount := 0
	for i := range gi.game.Factions {
		alive := false
		for x := range gi.game.Map.Grid {
			for y := range gi.game.Map.Grid[x] {
				if gi.game.Map.Grid[x][y].Owner == int8(i) &&
					gi.game.Map.Grid[x][y].Building == game.BuildingTownhall {
					alive = true
					break
				}
			}
			if alive {
				break
			}
		}
		gi.game.Factions[i].Alive = alive
		if alive {
			aliveCount++
		}
	}
	return aliveCount
}

func (gi *GameInstance) broadcastGameEnd() {
	type rankSortable struct {
		idx    int
		points int32
		alive  bool
		name   string
	}

	factions := gi.game.Factions[:]
	sorted := make([]rankSortable, 0, len(factions))
	for i := range factions {
		name := ""
		if factions[i].Player != nil {
			name = factions[i].Player.PlayerName
		}
		sorted = append(sorted, rankSortable{
			idx:    i,
			points: factions[i].Points,
			alive:  factions[i].Alive,
			name:   name,
		})
	}

	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].points > sorted[i].points {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	winnerFaction := -1
	winnerName := ""
	if len(sorted) > 0 {
		winnerFaction = sorted[0].idx
		winnerName = sorted[0].name
	}

	rankings := make([]packets.RankEntry, 0, len(sorted))
	for _, r := range sorted {
		rankings = append(rankings, packets.RankEntry{
			FactionIdx: r.idx,
			PlayerName: r.name,
			Points:     r.points,
			Alive:      r.alive,
		})
	}

	endPacket := &packets.S2CGameEndPacket{
		WinnerFaction: winnerFaction,
		WinnerName:    winnerName,
		Rankings:      rankings,
	}

	for _, c := range gi.clients {
		gi.sendToClient(c, endPacket)
	}
}
