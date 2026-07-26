package server

import (
	"fmt"
	"log"
	"math/rand"
	"sort"
	"sync"
	"time"

	gameai "github.com/threeidiotsonegamejam/gmtk26/src/ai"
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
	attackOrders   map[int][]game.AttackOrder
	// routePriorities records routes assigned this round so they receive
	// their promised first advancement before the regular round-robin queue.
	routePriorities    map[int]game.Hex
	actionResults      map[int]*game.ActionResult
	movementEvents     []game.MovementEvent
	attackEvents       []game.AttackEvent
	aiControllers      map[int]gameai.Planner
	pendingAITakeovers map[int]bool
	aiTraces           map[int]gameai.DecisionTrace
	clientsChanged     chan struct{}
	done               chan struct{}
	mu                 sync.RWMutex
}

func NewGameInstance(id uint64, g *game.Game, clients []*Client) *GameInstance {
	clientCount := min(len(clients), len((game.Game{}).Factions))
	clientSlots := append([]*Client(nil), clients[:clientCount]...)

	factionClients := make(map[*Client]int)
	for i, c := range clientSlots {
		if c != nil {
			factionClients[c] = i
		}
	}

	return &GameInstance{
		ID:                 id,
		game:               g,
		clients:            clientSlots,
		factionClients:     factionClients,
		actions:            make(map[int]*submittedAction),
		movementOrders:     make(map[int][]game.MovementOrder),
		attackOrders:       make(map[int][]game.AttackOrder),
		routePriorities:    make(map[int]game.Hex),
		actionResults:      make(map[int]*game.ActionResult),
		attackEvents:       nil,
		aiControllers:      make(map[int]gameai.Planner),
		pendingAITakeovers: make(map[int]bool),
		aiTraces:           make(map[int]gameai.DecisionTrace),
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
	if !gi.game.Factions[factionIdx].Alive {
		return fmt.Errorf("faction has been eliminated")
	}

	if actionType == game.ActionAttack {
		if attack == nil {
			return fmt.Errorf("attack payload was missing")
		}
		if !game.HexAdjacent(attack.From, attack.To) {
			return gi.setAttackOrderLocked(factionIdx, *attack)
		}
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
		!source.HasUnits() ||
		source.Units[0].Owner != factionOwner {
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
	gi.routePriorities[factionIdx] = move.From
	gi.attackOrders[factionIdx] = removeAttackOrder(gi.attackOrders[factionIdx], move.From)
	return nil
}

func (gi *GameInstance) setAttackOrderLocked(
	factionIdx int,
	attack game.AttackActionPayload,
) error {
	factionOwner := int8(factionIdx)
	source := gi.game.Map.GetCell(attack.From)
	if source == nil ||
		!source.HasUnits() ||
		source.Units[0].Owner != factionOwner ||
		source.Units[0].Type == game.UnitScout {
		return fmt.Errorf("no friendly non-Scout unit at attack source")
	}
	if gi.game.Map.GetCell(attack.To) == nil {
		return fmt.Errorf("attack target cell does not exist")
	}
	if game.HexAdjacent(attack.From, attack.To) {
		return fmt.Errorf("persistent attack target must not be adjacent")
	}

	gi.movementOrders[factionIdx] = removeMovementOrder(
		gi.movementOrders[factionIdx],
		attack.From,
	)
	gi.routePriorities[factionIdx] = attack.From
	gi.attackOrders[factionIdx] = removeAttackOrder(
		gi.attackOrders[factionIdx],
		attack.From,
	)
	gi.attackOrders[factionIdx] = append(
		gi.attackOrders[factionIdx],
		game.AttackOrder{From: attack.From, TargetTile: attack.To},
	)
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
	if priority, ok := gi.routePriorities[factionIdx]; ok && priority == from {
		delete(gi.routePriorities, factionIdx)
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
			f.Coins = game.StartingCoins
			f.Points = 0
			f.Resources = make(game.Resources)
			f.Alive = true
			if gi.clients[i] == nil {
				f.AI = true
			}
		}
	}

	gi.assignStartingCells()
	gi.initializeAIControllers()

	gameEndTime := time.Now().Add(game.MatchDuration)
	gi.game.GameEndTime = gameEndTime.UnixNano()
	firstDeadline := time.Now().Add(game.RoundDuration)

	gi.setRound(1)

	for i, c := range gi.clients {
		if c == nil {
			continue
		}
		f := gi.game.Factions[i]
		startPacket := &packets.S2CGameStartPacket{
			FactionIdx:   i,
			Map:          gi.game.Map,
			Coins:        f.Coins,
			Points:       f.Points,
			Resources:    f.Resources,
			Round:        1,
			Deadline:     firstDeadline.UnixNano(),
			GameEndTime:  gi.game.GameEndTime,
			Orders:       []game.MovementOrder{},
			AttackOrders: []game.AttackOrder{},
		}
		gi.sendToClient(c, startPacket)
	}

	for {
		if !gi.hasConnectedPlayers() {
			return
		}

		roundStart := time.Now()
		deadline := roundStart.Add(game.RoundDuration)

		for i, c := range gi.clients {
			if c == nil {
				continue
			}
			f := gi.game.Factions[i]
			gi.mu.RLock()
			orders := append([]game.MovementOrder{}, gi.movementOrders[i]...)
			attackOrders := append([]game.AttackOrder{}, gi.attackOrders[i]...)
			var result *game.ActionResult
			if gi.actionResults[i] != nil {
				copy := *gi.actionResults[i]
				result = &copy
			}
			movements := copyMovementEvents(gi.movementEvents)
			attackEvents := copyAttackEvents(gi.attackEvents)
			gi.mu.RUnlock()
			statePacket := &packets.S2CGameStatePacket{
				Round:        gi.game.Round,
				Deadline:     deadline.UnixNano(),
				Map:          gi.game.Map,
				Coins:        f.Coins,
				Points:       f.Points,
				Resources:    f.Resources,
				Orders:       orders,
				Result:       result,
				Movements:    movements,
				AttackOrders: attackOrders,
				AttackEvents: attackEvents,
			}
			gi.sendToClient(c, statePacket)
		}

		if !gi.waitUntil(deadline) {
			return
		}

		gi.mu.Lock()
		aliveCount := gi.resolveRoundLocked()
		gi.mu.Unlock()

		if time.Now().After(gameEndTime) || aliveCount <= 1 {
			gi.broadcastGameEnd()
			return
		}
	}
}

func copyAttackEvents(events []game.AttackEvent) []game.AttackEvent {
	copied := make([]game.AttackEvent, len(events))
	copy(copied, events)
	return copied
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

	island := m.LargestLandIsland()
	candidates := make([]startCandidate, 0)
	for x := range m.Grid {
		for y := range m.Grid[x] {
			hex := game.NewHex(int32(x), int32(y))
			if m.Grid[x][y].Tile == game.TilePlains && island[hex] {
				quality := startingZoneQuality(m, hex)
				candidates = append(candidates, startCandidate{
					hex:       hex,
					quality:   quality,
					qualified: quality == startZoneRequirementCount,
				})
			}
		}
	}

	// A pathological generated map may contain no Plains on its largest
	// island. Preserve a playable fallback while keeping normal starts on
	// Plains.
	if len(candidates) == 0 {
		for x := range m.Grid {
			for y := range m.Grid[x] {
				hex := game.NewHex(int32(x), int32(y))
				if island[hex] {
					candidates = append(candidates, startCandidate{hex: hex})
				}
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

	claimed := make([]game.Hex, 0, len(gi.game.Factions))
	used := make(map[game.Hex]bool)

	for i := range gi.game.Factions {
		hex, ok := pickStartingHex(candidates, claimed, used)
		if !ok {
			break
		}
		used[hex] = true
		claimed = append(claimed, hex)
		cell := m.GetCell(hex)
		cell.Owner = int8(i)
		cell.Building = &game.BuildingData{Type: game.BuildingTownhall, HP: game.BuildingMaxHP(game.BuildingTownhall)}
	}
}

const (
	startZoneRadius           = int32(8)
	startZoneMinimumSpacing   = int32(14)
	startZoneRequirementCount = 5
)

type startCandidate struct {
	hex       game.Hex
	quality   int
	qualified bool
}

func startingZoneQuality(m *game.Map, center game.Hex) int {
	hasFarm := false
	hasForester := false
	hasRock := false
	hasIron := false
	plains := 0

	minX := max(int32(0), center.X-startZoneRadius)
	maxX := min(m.GridSize.X-1, center.X+startZoneRadius)
	minY := max(int32(0), center.Y-startZoneRadius)
	maxY := min(m.GridSize.Y-1, center.Y+startZoneRadius)
	for x := minX; x <= maxX; x++ {
		for y := minY; y <= maxY; y++ {
			hex := game.NewHex(x, y)
			if center.Distance(hex) > startZoneRadius {
				continue
			}
			cell := m.GetCell(hex)
			if cell == nil {
				continue
			}
			switch cell.Tile {
			case game.TilePlains:
				plains++
				if game.BuildingCanPlace(m, game.BuildingFarm, hex) {
					hasFarm = true
				}
			case game.TileForest, game.TileJungle:
				hasForester = true
			case game.TileRock:
				hasRock = true
			case game.TileIron:
				hasIron = true
			}
		}
	}

	quality := 0
	for _, met := range []bool{hasFarm, hasForester, hasRock, hasIron, plains >= 3} {
		if met {
			quality++
		}
	}
	return quality
}

func pickStartingHex(
	candidates []startCandidate,
	claimed []game.Hex,
	used map[game.Hex]bool,
) (game.Hex, bool) {
	if len(claimed) == 0 {
		best := -1
		for i, candidate := range candidates {
			if used[candidate.hex] {
				continue
			}
			if best < 0 ||
				candidate.qualified && !candidates[best].qualified ||
				candidate.qualified == candidates[best].qualified &&
					candidate.quality > candidates[best].quality {
				best = i
			}
		}
		if best >= 0 {
			return candidates[best].hex, true
		}
		return game.Hex{}, false
	}

	for spacing := startZoneMinimumSpacing; spacing >= 0; spacing-- {
		best := -1
		var bestDistance int32
		for i, candidate := range candidates {
			if used[candidate.hex] {
				continue
			}
			minDistance := candidate.hex.Distance(claimed[0])
			for _, other := range claimed[1:] {
				minDistance = min(minDistance, candidate.hex.Distance(other))
			}
			if minDistance < spacing {
				continue
			}
			if best < 0 ||
				candidate.qualified && !candidates[best].qualified ||
				candidate.qualified == candidates[best].qualified && minDistance > bestDistance ||
				candidate.qualified == candidates[best].qualified && minDistance == bestDistance &&
					candidate.quality > candidates[best].quality {
				best = i
				bestDistance = minDistance
			}
		}
		if best >= 0 {
			return candidates[best].hex, true
		}
	}
	return game.Hex{}, false
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
func (gi *GameInstance) clientLeft(client *Client) {
	gi.mu.Lock()
	if factionIdx, ok := gi.factionClients[client]; ok &&
		!gi.game.Factions[factionIdx].AI {
		gi.pendingAITakeovers[factionIdx] = true
	}
	gi.mu.Unlock()

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
		faction := &gi.game.Factions[i]
		if !faction.Alive {
			continue
		}
		coins, resources := game.ResolveFactionRoundIncome(
			&gi.game.Map,
			int8(i),
			faction.Resources,
		)
		faction.Resources = resources
		faction.Coins += coins
	}
}

func (gi *GameInstance) awardControlScore() {
	for i := range gi.game.Factions {
		faction := &gi.game.Factions[i]
		if !faction.Alive {
			continue
		}
		faction.Points += game.FactionControlScore(&gi.game.Map, int8(i))
	}
}

func controlScoreDue(round int32) bool {
	return round > 0 && round%game.ScoreIntervalRounds == 0
}

func (gi *GameInstance) checkAlive() int {
	aliveCount := 0
	for i := range gi.game.Factions {
		alive := false
		for x := range gi.game.Map.Grid {
			for y := range gi.game.Map.Grid[x] {
				if gi.game.Map.Grid[x][y].Owner == int8(i) &&
					gi.game.Map.Grid[x][y].BuildingType() == game.BuildingTownhall {
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
			continue
		}
		delete(gi.actions, i)
		delete(gi.movementOrders, i)
		delete(gi.attackOrders, i)
		delete(gi.routePriorities, i)
		delete(gi.aiControllers, i)
		delete(gi.aiTraces, i)
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

	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].alive != sorted[j].alive {
			return sorted[i].alive
		}
		if sorted[i].points != sorted[j].points {
			return sorted[i].points > sorted[j].points
		}
		return sorted[i].idx < sorted[j].idx
	})

	winnerFaction := -1
	winnerName := ""
	eligible := make([]rankSortable, 0, len(sorted))
	for _, faction := range sorted {
		if faction.alive {
			eligible = append(eligible, faction)
		}
	}
	if len(eligible) == 0 {
		eligible = sorted
	}
	if len(eligible) == 1 ||
		len(eligible) > 1 && eligible[0].points > eligible[1].points {
		winnerFaction = eligible[0].idx
		winnerName = eligible[0].name
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
