package server

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/net/packets"
)

type submittedAction struct {
	Type    game.ActionType
	Build   *game.BuildActionPayload
	Recruit *game.RecruitActionPayload
	Attack  *game.AttackActionPayload
}

type GameInstance struct {
	ID             uint64
	game           *game.Game
	clients        []*Client
	factionClients map[*Client]int
	actions        map[int]*submittedAction
	movementOrders map[int][]game.MovementOrder
	// movementPriorities records routes assigned this round so they receive
	// their promised first advancement before the regular round-robin queue.
	movementPriorities map[int]game.Hex
	actionResults      map[int]*game.ActionResult
	movementEvents     []game.MovementEvent
	clientsChanged     chan struct{}
	done               chan struct{}
	mu                 sync.RWMutex
}

func NewGameInstance(id uint64, g *game.Game, clients []*Client) *GameInstance {
	factionClients := make(map[*Client]int)
	for i, c := range clients {
		if c != nil {
			factionClients[c] = i
		}
	}

	return &GameInstance{
		ID:                 id,
		game:               g,
		clients:            clients,
		factionClients:     factionClients,
		actions:            make(map[int]*submittedAction),
		movementOrders:     make(map[int][]game.MovementOrder),
		movementPriorities: make(map[int]game.Hex),
		actionResults:      make(map[int]*game.ActionResult),
		clientsChanged:     make(chan struct{}, 1),
		done:               make(chan struct{}),
	}
}

func (gi *GameInstance) SubmitAction(
	c *Client,
	round int32,
	actionType game.ActionType,
	build *game.BuildActionPayload,
	move *game.MoveActionPayload,
	recruit *game.RecruitActionPayload,
	attack *game.AttackActionPayload,
) error {
	gi.mu.Lock()
	defer gi.mu.Unlock()

	if round != gi.game.Round {
		return fmt.Errorf("wrong round: got %d, want %d", round, gi.game.Round)
	}

	factionIdx, ok := gi.factionClients[c]
	if !ok {
		return fmt.Errorf("client not in this game")
	}

	if actionType == game.ActionMove {
		if move == nil {
			return fmt.Errorf("move payload was missing")
		}
		return gi.setMovementOrderLocked(factionIdx, *move)
	}

	gi.actions[factionIdx] = &submittedAction{
		Type:    actionType,
		Build:   build,
		Recruit: recruit,
		Attack:  attack,
	}
	return nil
}

func (gi *GameInstance) setMovementOrderLocked(
	factionIdx int,
	move game.MoveActionPayload,
) error {
	factionOwner := int8(factionIdx)
	source := gi.game.Map.GetCell(move.From)
	if source == nil ||
		source.Unit == game.UnitUnknown ||
		source.UnitOwner != factionOwner {
		return fmt.Errorf("no friendly unit at movement source")
	}
	if move.From == move.To {
		return fmt.Errorf("movement destination must differ from source")
	}
	if _, ok := gi.game.Map.FindUnitPath(factionOwner, move.From, move.To); !ok {
		return fmt.Errorf("no legal route to destination")
	}

	orders := removeMovementOrder(gi.movementOrders[factionIdx], move.From)
	gi.movementOrders[factionIdx] = append(orders, game.MovementOrder{
		Current:     move.From,
		Destination: move.To,
	})
	gi.movementPriorities[factionIdx] = move.From
	return nil
}

// CancelMovementOrder is an immediate, free command. It also withdraws a
// newly assigned priority so cancelling never advances it at the boundary.
func (gi *GameInstance) CancelMovementOrder(c *Client, round int32, from game.Hex) error {
	gi.mu.Lock()
	defer gi.mu.Unlock()

	if round != gi.game.Round {
		return fmt.Errorf("wrong round: got %d, want %d", round, gi.game.Round)
	}
	factionIdx, ok := gi.factionClients[c]
	if !ok {
		return fmt.Errorf("client not in this game")
	}

	gi.movementOrders[factionIdx] = removeMovementOrder(
		gi.movementOrders[factionIdx],
		from,
	)
	if priority, ok := gi.movementPriorities[factionIdx]; ok && priority == from {
		delete(gi.movementPriorities, factionIdx)
	}
	return nil
}

// CancelBuildAction withdraws only the pending build at the supplied target.
// A stale cancel therefore cannot erase a newer replacement action.
func (gi *GameInstance) CancelBuildAction(c *Client, round int32, to game.Hex) error {
	gi.mu.Lock()
	defer gi.mu.Unlock()

	if round != gi.game.Round {
		return fmt.Errorf("wrong round: got %d, want %d", round, gi.game.Round)
	}
	factionIdx, ok := gi.factionClients[c]
	if !ok {
		return fmt.Errorf("client not in this game")
	}

	action := gi.actions[factionIdx]
	if action != nil &&
		action.Type == game.ActionBuild &&
		action.Build != nil &&
		action.Build.To == to {
		delete(gi.actions, factionIdx)
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
	gi.game.GameEndTime = gameEndTime.UnixNano()
	firstDeadline := time.Now().Add(5 * time.Second)

	gi.setRound(1)

	for i, c := range gi.clients {
		if c == nil {
			continue
		}
		f := gi.game.Factions[i]
		startPacket := &packets.S2CGameStartPacket{
			FactionIdx:  i,
			Map:         gi.game.Map,
			Coins:       f.Coins,
			Points:      f.Points,
			Resources:   f.Resources,
			Round:       1,
			Deadline:    firstDeadline.UnixNano(),
			GameEndTime: gi.game.GameEndTime,
			Orders:      []game.MovementOrder{},
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
			gi.mu.RLock()
			orders := append([]game.MovementOrder{}, gi.movementOrders[i]...)
			var result *game.ActionResult
			if gi.actionResults[i] != nil {
				copy := *gi.actionResults[i]
				result = &copy
			}
			movements := copyMovementEvents(gi.movementEvents)
			gi.mu.RUnlock()
			statePacket := &packets.S2CGameStatePacket{
				Round:     gi.game.Round,
				Deadline:  deadline.UnixNano(),
				Map:       gi.game.Map,
				Coins:     f.Coins,
				Points:    f.Points,
				Resources: f.Resources,
				Orders:    orders,
				Result:    result,
				Movements: movements,
			}
			gi.sendToClient(c, statePacket)
		}

		if !gi.waitUntil(deadline) {
			return
		}

		gi.mu.Lock()
		gi.processAutoActions()
		gi.processClientActions()
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

func copyMovementEvents(events []game.MovementEvent) []game.MovementEvent {
	copied := make([]game.MovementEvent, len(events))
	for i, event := range events {
		copied[i] = event
		copied[i].Path = append([]game.Hex(nil), event.Path...)
	}
	return copied
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

	// A surviving faction wins an elimination game even when an eliminated
	// faction accumulated more points. When time expires (or everyone is
	// eliminated), standings and the winner are decided by score.
	survivingFaction := -1
	for _, faction := range sorted {
		if !faction.alive {
			continue
		}
		if survivingFaction >= 0 {
			survivingFaction = -1
			break
		}
		survivingFaction = faction.idx
	}
	sort.SliceStable(sorted, func(i, j int) bool {
		if survivingFaction >= 0 {
			iSurvived := sorted[i].idx == survivingFaction
			jSurvived := sorted[j].idx == survivingFaction
			if iSurvived != jSurvived {
				return iSurvived
			}
		}
		if sorted[i].points != sorted[j].points {
			return sorted[i].points > sorted[j].points
		}
		return sorted[i].idx < sorted[j].idx
	})

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
