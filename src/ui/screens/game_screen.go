package screens

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/audio"
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	"github.com/threeidiotsonegamejam/gmtk26/src/global"
	gameNet "github.com/threeidiotsonegamejam/gmtk26/src/net"
	"github.com/threeidiotsonegamejam/gmtk26/src/net/packets"
	"github.com/threeidiotsonegamejam/gmtk26/src/render"
	"github.com/threeidiotsonegamejam/gmtk26/src/settings"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui/anchor"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

const gameCodeCopyFeedbackDuration = 1500 * time.Millisecond

var gameSeedInput = ui.Input()
var gameWorld = ui.GameWorld()
var gameRegenerateButton = ui.Button()
var currentGame *game.Game
var gamePreviousScreen *ui.ScreenElement
var gameLeaveTransition bool

// localClientID is the server-assigned identity from the connect handshake,
// used to decide whether the local player hosts the current game.
var localClientID game.ClientID

// serverGameActive is true between S2CGameStartPacket and S2CGameEndPacket
// for both remote multiplayer and the in-process solo server.
var serverGameActive bool
var serverRound int32
var serverCoins int32
var serverPoints int32
var serverResources game.Resources
var serverDeadline int64
var serverGameEndTime int64
var roundAnnouncement int32
var roundAnnouncementUntil time.Time
var focusTownhallPending bool
var gameOverMessage string
var gameActionError string
var gamePendingAction string
var gameResolutionMessage string
var gameResolutionUntil time.Time
var gameLastResolutionKey string

const (
	roundCountdownDuration    = 3 * time.Second
	roundAnnouncementDuration = time.Second
)

func init() {
	// In a running authoritative game, building placement goes through the
	// server instead of mutating the local map; the next state broadcast
	// carries the result.
	gameWorld.Renderer.OnPlaceBuilding = func(from, to game.Hex, building game.BuildingType) bool {
		if !serverGameActive {
			return false
		}
		if err := gameNet.SendBuildAction(serverRound, from, to, building); err != nil {
			fmt.Printf("failed to send build action: %v\n", err)
		} else {
			// Pending actions are intentionally client-only: opponents should
			// not see a building until the server resolves the round.
			gameWorld.Renderer.QueueBuilding(to, building)
			setPendingAction(fmt.Sprintf("Build %s", building))
		}
		return true
	}
	gameWorld.Renderer.OnRecruit = func(from, to game.Hex, unit game.UnitType) bool {
		if !serverGameActive {
			return false
		}
		if err := gameNet.SendRecruitAction(serverRound, from, to, unit); err != nil {
			fmt.Printf("failed to send recruit action: %v\n", err)
			return true
		}
		gameWorld.Renderer.ClearQueuedBuilding()
		setPendingAction(fmt.Sprintf("Recruit %s", unit))
		return true
	}
	gameWorld.Renderer.OnMove = func(from, to game.Hex) bool {
		if !serverGameActive {
			return false
		}
		if err := gameNet.SendMoveAction(serverRound, from, to); err != nil {
			fmt.Printf("failed to send move action: %v\n", err)
			return false
		}
		return true
	}
	gameWorld.Renderer.OnAttack = func(from, to game.Hex) bool {
		if !serverGameActive {
			return false
		}
		if err := gameNet.SendAttackAction(serverRound, from, to); err != nil {
			fmt.Printf("failed to send attack action: %v\n", err)
			return true
		}
		setPendingAction("Attack ordered")
		return true
	}
	gameWorld.Renderer.OnCancelMovement = func(from game.Hex) bool {
		if !serverGameActive {
			return false
		}
		if err := gameNet.SendCancelMovementOrder(serverRound, from); err != nil {
			fmt.Printf("failed to cancel movement order: %v\n", err)
			return false
		}
		showResolutionToast("Movement order cancelled", fmt.Sprintf("cancel:%d:%d:%d", serverRound, from.X, from.Y))
		return true
	}
	gameWorld.Renderer.OnCancelBuilding = func(to game.Hex) bool {
		if !serverGameActive {
			return false
		}
		if err := gameNet.SendCancelBuildAction(serverRound, to); err != nil {
			fmt.Printf("failed to cancel build action: %v\n", err)
			return false
		}
		gamePendingAction = ""
		showResolutionToast("Pending build cancelled", fmt.Sprintf("cancel-build:%d:%d:%d", serverRound, to.X, to.Y))
		return true
	}
}

func setPendingAction(label string) {
	gamePendingAction = label
}

func SetLocalClientID(id game.ClientID) {
	localClientID = id
}

func isLocalHost() bool {
	return currentGame != nil && currentGame.HostID == localClientID
}

// ApplyServerGameStart switches the pre-game session into the running game
// using the authoritative state from the server.
func ApplyServerGameStart(p *packets.S2CGameStartPacket) {
	if currentGame == nil {
		return
	}
	serverGameActive = true
	gameOverMessage = ""
	clearRoundAnnouncement()
	gameWorld.Renderer.ClearQueuedBuilding()
	gameWorld.Renderer.LocalFaction = int8(p.FactionIdx)
	gameWorld.Renderer.ActionsEnabled = true
	applyServerRound(p.Round, p.Deadline, p.Map, p.Coins, p.Points, p.Resources, p.Orders, p.AttackOrders, nil, nil, nil)
	serverGameEndTime = p.GameEndTime
	// The renderer may not be initialized until the game screen's first
	// update. Defer focus so ResetCamera cannot discard this request.
	focusTownhallPending = true
}

func ApplyServerGameState(p *packets.S2CGameStatePacket) {
	if !serverGameActive {
		return
	}
	if p.Round != serverRound {
		gameWorld.Renderer.ClearQueuedBuilding()
		roundAnnouncement = p.Round
		roundAnnouncementUntil = time.Now().Add(roundAnnouncementDuration)
		gamePendingAction = ""
	}
	applyServerRound(p.Round, p.Deadline, p.Map, p.Coins, p.Points, p.Resources, p.Orders, p.AttackOrders, p.Movements, p.AttackEvents, p.Result)
}

func ApplyServerGameEnd(p *packets.S2CGameEndPacket) {
	if !serverGameActive {
		return
	}
	serverGameActive = false
	serverGameEndTime = 0
	clearRoundAnnouncement()
	focusTownhallPending = false
	gameWorld.Renderer.ActionsEnabled = false
	gameWorld.Renderer.ClearQueuedBuilding()
	gameOverMessage = "Game over!"

	// Keep the authoritative result independent of the network packet buffer
	// and transition away from the live world into a dedicated results screen.
	result := *p
	result.Rankings = append([]packets.RankEntry(nil), p.Rankings...)
	if gameNet.LocalGameActive() {
		// The solo transport otherwise remains selected after its server game
		// finishes, preventing the next solo game from being created.
		gameNet.StopLocalGame()
	}
	SetActiveScreen(NewGameEndScreen(
		result,
		gameNet.LocalGameState.FactionIdx,
		gamePreviousScreen,
	))
}

func applyServerRound(
	round int32,
	deadline int64,
	m game.Map,
	coins, points int32,
	resources game.Resources,
	orders []game.MovementOrder,
	attackOrders []game.AttackOrder,
	movements []game.MovementEvent,
	attackEvents []game.AttackEvent,
	result *game.ActionResult,
) {
	if gameWorld.Renderer.SelectedKind == render.SelectionUnit &&
		gameWorld.Renderer.SelectedHex != nil {
		for _, movement := range movements {
			if movement.Owner != int8(gameNet.LocalGameState.FactionIdx) || len(movement.Path) < 2 {
				continue
			}
			if movement.Path[0] == *gameWorld.Renderer.SelectedHex {
				endpoint := movement.Path[len(movement.Path)-1]
				gameWorld.Renderer.SelectedHex = &endpoint
				break
			}
		}
	}
	serverRound = round
	serverDeadline = deadline
	serverCoins = coins
	serverPoints = points
	serverResources = resources
	gameWorld.Renderer.LocalCoins = coins
	gameWorld.Renderer.LocalResources = resources
	currentGame.Round = round
	currentGame.Map = m
	gameWorld.Map = m
	gameWorld.Renderer.SetMovementOrders(orders)
	gameWorld.Renderer.SetAttackOrders(attackOrders)
	gameWorld.Renderer.StartMovementAnimations(movements)
	gameWorld.Renderer.StartAttackAnimations(attackEvents)
	if result != nil {
		key := fmt.Sprintf("%d:%d:%d:%t", result.Round, result.Type, result.Status, result.Automatic)
		showResolutionToast(result.Message, key)
	}
	gameSeedInput.SetText(strconv.FormatInt(m.Seed, 10))
}

func showResolutionToast(message, key string) {
	if message == "" || key == gameLastResolutionKey {
		return
	}
	gameLastResolutionKey = key
	gameResolutionMessage = message
	gameResolutionUntil = time.Now().Add(2500 * time.Millisecond)
}

func multiplayerStatusText() string {
	if currentGame == nil {
		return ""
	}
	if gameOverMessage != "" {
		return gameOverMessage
	}
	if serverGameActive {
		remaining := max(time.Until(time.Unix(0, serverDeadline)), 0)
		return fmt.Sprintf(
			"Round %d | Coins: %d | Points: %d | Next round: %ds",
			serverRound,
			serverCoins,
			serverPoints,
			countdownSecondsRemaining(remaining),
		)
	}
	if !currentGame.Multiplayer {
		return "Starting solo game..."
	}

	players := 0
	for i := range currentGame.Factions {
		if currentGame.Factions[i].Player != nil {
			players++
		}
	}
	if isLocalHost() {
		return fmt.Sprintf("%d/%d players joined", players, currentGame.MaxPlayers)
	}
	return fmt.Sprintf(
		"%d/%d players | Waiting for the host to start",
		players,
		currentGame.MaxPlayers,
	)
}

func countdownSecondsRemaining(remaining time.Duration) int {
	if remaining <= 0 {
		return 0
	}
	return int(math.Ceil(float64(remaining) / float64(time.Second)))
}

// roundCountdownState maps the authoritative time remaining to one second for
// each digit. Progress restarts at zero when the digit changes.
func roundCountdownState(remaining time.Duration) (digit int, progress float64, visible bool) {
	if remaining <= 0 || remaining > roundCountdownDuration {
		return 0, 0, false
	}

	seconds := float64(remaining) / float64(time.Second)
	digit = countdownSecondsRemaining(remaining)
	progress = min(max(float64(digit)-seconds, 0), 1)
	return digit, progress, true
}

func easeInExpo(progress float64) float64 {
	progress = min(max(progress, 0), 1)
	switch progress {
	case 0:
		return 0
	case 1:
		return 1
	default:
		return math.Pow(2, 10*progress-10)
	}
}

func roundCountdownTextSize(progress float64, renderHeight int32, scale float32) int32 {
	targetSize := float64(min(max(renderHeight/3, int32(180)), int32(320))) * float64(scale)
	startSize := targetSize * 0.4
	size := startSize + (targetSize-startSize)*easeInExpo(progress)
	return int32(math.Round(size))
}

func roundAnnouncementState(round int32, remaining time.Duration) (text string, visible bool) {
	if round <= 0 || remaining <= 0 {
		return "", false
	}
	return fmt.Sprintf("Round #%d", round), true
}

func roundAnnouncementTextSize(renderHeight int32, scale float32) int32 {
	baseSize := min(max(renderHeight/8, int32(72)), int32(128))
	return int32(math.Round(float64(baseSize) * float64(scale)))
}

func roundCountdownPosition(value settings.CountdownAnchor, renderSize vec.Vec2i) vec.Vec2i {
	column, row, valid := settings.CountdownAnchorGridPosition(value)
	if !valid {
		column, row, _ = settings.CountdownAnchorGridPosition(settings.CountdownCenter)
	}
	denominator := settings.CountdownGridSize * 2
	return vec.Vec2i{
		X: renderSize.X * (column*2 + 1) / denominator,
		Y: renderSize.Y * (row*2 + 1) / denominator,
	}
}

func clearRoundAnnouncement() {
	roundAnnouncement = 0
	roundAnnouncementUntil = time.Time{}
}

func newRoundCountdown() *ui.GroupElement {
	displayText := ""
	visible := false

	text := ui.Text().
		WithTextDynamic(func() string {
			return displayText
		}).
		WithTextSize(roundCountdownTextSize(
			0,
			int32(rl.GetRenderHeight()),
			settings.Current.CountdownScale,
		)).
		WithTextColor(color.RGBA{R: 255, G: 244, B: 194, A: 255}).
		WithTextShadow(color.RGBA{R: 14, G: 12, B: 25, A: 220}, vec.Vec2i{X: 7, Y: 9}).
		WithAnchors(anchor.Center, anchor.Center).
		WithVisibleDynamic(func(el *ui.TextElement) bool {
			return visible
		})

	return ui.Group().
		WithAnchors(anchor.Center, anchor.TopLeft).
		WithRelativePosDynamic(func(el *ui.GroupElement) vec.Vec2i {
			return roundCountdownPosition(
				settings.Current.CountdownAnchor,
				el.Parent.Size(),
			)
		}).
		WithUpdate(func(deltaNano int64) {
			if !serverGameActive {
				visible = false
				return
			}

			now := time.Now()
			var progress float64
			digit, progress, countdownVisible := roundCountdownState(
				time.Unix(0, serverDeadline).Sub(now),
			)
			if countdownVisible {
				displayText = strconv.Itoa(digit)
				visible = true
				text.TextSize = roundCountdownTextSize(
					progress,
					int32(rl.GetRenderHeight()),
					settings.Current.CountdownScale,
				)
				return
			}

			displayText, visible = roundAnnouncementState(
				roundAnnouncement,
				roundAnnouncementUntil.Sub(now),
			)
			if visible {
				text.TextSize = roundAnnouncementTextSize(
					int32(rl.GetRenderHeight()),
					settings.Current.CountdownScale,
				)
			}
		}).
		AddChild(text)
}

func EnterGame(state game.Game) {
	clearMatchmaking()
	serverGameActive = false
	clearRoundAnnouncement()
	focusTownhallPending = false
	gameOverMessage = ""
	gameActionError = ""
	gameWorld.Renderer.ClearQueuedBuilding()
	gameWorld.Renderer.SetMovementOrders(nil)
	gameWorld.Renderer.ClearSelection()
	gameWorld.Renderer.ActionsEnabled = false
	applyGameState(state)
	gameWorld.Renderer.ResetCamera(&gameWorld.Map)
	if screenIsActiveOrPending(gameScreen) {
		return
	}
	// Return to the play menu on close, not intermediate creation/matchmaking screens.
	SetActiveScreen(NewGameScreen(playScreenOrNew()))
}

func RejectGameStart(message string) {
	gameActionError = capitalizeSentence(message)
}

func capitalizeSentence(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return s
	}
	return string(unicode.ToUpper(r)) + s[size:]
}

func RejectGameJoin(message string) {
	clearMatchmaking()
	SetPlayError(message)
	SetActiveScreen(playScreenOrNew())
}

func ApplyGameUpdate(state game.Game) {
	if currentGame == nil || currentGame.GameID != state.GameID {
		return
	}
	applyGameState(state)
}

func CloseGame(gameID uint64) {
	if currentGame == nil || currentGame.GameID != gameID {
		return
	}
	if gameLeaveTransition && activeScreen == gameScreen && pendingScreen != nil {
		// The local leave action already owns the transition and destination.
		// Keep its outgoing game frame intact until the crossfade completes.
		return
	}
	clearCurrentGame()
	GoToPreviousScreen(gamePreviousScreen)
}

func ResetGameSession() {
	wasMatchmaking := matchmakingActive
	clearMatchmaking()
	if currentGame == nil || !currentGame.Multiplayer {
		if wasMatchmaking {
			SetActiveScreen(playScreenOrNew())
		}
		return
	}
	clearCurrentGame()
	GoToPreviousScreen(gamePreviousScreen)
}

func LeaveCurrentGame() {
	if matchmakingActive {
		if err := gameNet.Send(&packets.C2SLeaveGamePacket{}); err != nil {
			fmt.Printf("failed to cancel matchmaking: %v\n", err)
		}
		clearMatchmaking()
	}
	if currentGame == nil {
		return
	}
	gameLeaveTransition = true
	if gameNet.LocalGameActive() {
		gameNet.StopLocalGame()
	} else if currentGame.Multiplayer {
		if err := gameNet.Send(&packets.C2SLeaveGamePacket{}); err != nil {
			println("failed to leave game:", err.Error())
		}
	}
}

func applyGameState(state game.Game) {
	nextMap := state.Map
	if currentGame != nil &&
		currentGame.GameID == state.GameID &&
		gameWorld.Map.Seed == nextMap.Seed &&
		nextMap.Grid == nil {
		nextMap = gameWorld.Map
	}
	if nextMap.Grid == nil {
		nextMap.Generate()
	}

	state.Map = nextMap
	gameWorld.Map = nextMap
	gameSeedInput.SetText(strconv.FormatInt(nextMap.Seed, 10))
	currentGame = &state
}

func clearCurrentGame() {
	currentGame = nil
	gameWorld.Map = game.Map{}
	gameWorld.Renderer.ClearQueuedBuilding()
	gameWorld.Renderer.ClearSelection()
	gameLeaveTransition = false
	serverGameActive = false
	serverGameEndTime = 0
	clearRoundAnnouncement()
	focusTownhallPending = false
	gameOverMessage = ""
	gameActionError = ""
	gamePendingAction = ""
	gameResolutionMessage = ""
	gameLastResolutionKey = ""
	serverResources = nil
	gameNet.LocalGameState.Reset()
}

func setBuildingClick(building game.BuildingType) func() {
	return func() {
		gameWorld.Renderer.BuildingToPlace = building
		gameWorld.Renderer.RecruitToPlace = game.UnitUnknown
	}
}

func buildingButtonTooltip(building game.BuildingType) string {
	cost := buildingCostText(building)
	score := game.BuildingControlScore(building, game.TileRock)
	switch building {
	case game.BuildingBarracks:
		return fmt.Sprintf(
			"Cost: %s\nRecruits: Peasant, Archer, Knight, Scout\nPlace on: Plains\nControl: +%d every 30s",
			cost, score,
		)
	case game.BuildingFarm:
		return fmt.Sprintf(
			"Cost: %s\nProduces: Food +1\nPlace on: Plains adjacent to Water\nControl: +%d every 30s",
			cost, score,
		)
	case game.BuildingMine:
		return fmt.Sprintf(
			"Cost: %s\nProduces: Rock -> Stone +1, Iron -> Iron +1, Gold -> 1 Coin + 1 Gold\nPlace on: Rock, Iron, or Gold\nControl: +3 (Gold +5) every 30s",
			cost,
		)
	case game.BuildingForester:
		return fmt.Sprintf(
			"Cost: %s\nProduces: Wood +1\nPlace on: Forest or Jungle\nControl: +%d every 30s",
			cost, score,
		)
	case game.BuildingBank:
		return fmt.Sprintf(
			"Cost: %s\nConverts: 5 Gold -> 5 Coins each round\nPlace on: Plains\nControl: +%d every 30s",
			cost, score,
		)
	default:
		return ""
	}
}

func buildingCostText(building game.BuildingType) string {
	parts := make([]string, 0, 3)
	if coins := game.BuildingCost(building); coins > 0 {
		parts = append(parts, fmt.Sprintf("%d Coins", coins))
	}
	costs := game.BuildingResourceCost(building)
	for _, resource := range []game.ResourceType{
		game.ResourceFood,
		game.ResourceWood,
		game.ResourceStone,
		game.ResourceIron,
		game.ResourceGold,
	} {
		if amount := costs[resource]; amount > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", amount, resource))
		}
	}
	return strings.Join(parts, ", ")
}

func tryFocusOnTownhall() bool {
	for x := range gameWorld.Map.Grid {
		for y := range gameWorld.Map.Grid[x] {
			cell := &gameWorld.Map.Grid[x][y]
			if cell.Owner == int8(gameNet.LocalGameState.FactionIdx) && cell.BuildingType() == game.BuildingTownhall {
				gameWorld.Renderer.FocusOnHex(game.NewHex(int32(x), int32(y)))
				return true
			}
		}
	}
	return false
}

func focusOnTownhall() {
	tryFocusOnTownhall()
}

func HandleTownHallShortcut() {
	if activeScreen != gameScreen || IsTransitioning() || IsEscScreenOpen() || !serverGameActive {
		return
	}
	focusOnTownhall()
}

func buildingToolbar() *ui.GroupElement {
	group := ui.Group().
		WithAnchors(anchor.BottomLeft, anchor.BottomLeft).
		WithRelativePos(vec.Vec2i{X: 8, Y: -48}).
		WithVisibleDynamic(func(*ui.GroupElement) bool {
			if !serverGameActive ||
				gameWorld.Renderer.SelectedHex == nil ||
				gameWorld.Renderer.SelectedKind != render.SelectionUnit {
				return false
			}
			cell := gameWorld.Map.GetCell(*gameWorld.Renderer.SelectedHex)
			return cell != nil &&
				cell.HasUnits() && cell.Units[0].Type == game.UnitScout &&
				cell.Units[0].Owner == int8(gameNet.LocalGameState.FactionIdx)
		}).
		AddChild(
			ui.Text().
				WithText("Buildings:").
				WithTextSize(24).
				WithTextColor(rl.White).
				WithRelativePos(vec.Vec2i{Y: -32}),
		)

	buttons := []struct {
		label    string
		building game.BuildingType
		x        int32
	}{
		{label: "None", building: game.BuildingUnknown, x: 0},
		{label: "Barracks", building: game.BuildingBarracks, x: 84},
		{label: "Farm", building: game.BuildingFarm, x: 224},
		{label: "Mine", building: game.BuildingMine, x: 308},
		{label: "Forester", building: game.BuildingForester, x: 384},
		{label: "Bank", building: game.BuildingBank, x: 520},
	}
	for _, item := range buttons {
		group.AddChild(
			ui.Button().
				WithText(item.label).
				WithTextSize(24).
				WithPadding(8).
				WithRelativePos(vec.Vec2i{X: item.x}).
				WithEnabledDynamic(func(*ui.ButtonElement) bool {
					if item.building == game.BuildingUnknown {
						return true
					}
					return game.CanAffordBuildingAfterRoundIncome(
						&gameWorld.Map,
						int8(gameNet.LocalGameState.FactionIdx),
						item.building,
						serverCoins,
						serverResources,
					)
				}).
				WithClick(setBuildingClick(item.building)).
				WithTooltip(buildingButtonTooltip(item.building)),
		)
	}
	return group
}

func serverResourceList() *ui.GroupElement {
	group := ui.Group().
		WithAnchors(anchor.TopLeft, anchor.TopLeft).
		WithRelativePos(vec.Vec2i{X: 8, Y: 8}).
		WithVisibleDynamic(func(*ui.GroupElement) bool {
			return serverGameActive
		})

	for i, resourceType := range serverResourceDisplayOrder {
		group.AddChild(
			ui.Text().
				WithTextDynamic(func() string {
					return fmt.Sprintf("%s: %d", resourceType.String(), serverResources[resourceType])
				}).
				WithTextSize(20).
				WithTextColor(rl.White).
				WithRelativePos(vec.Vec2i{Y: int32(i+1) * 24}),
		)
	}
	group.AddChild(
		ui.Button().
			WithText("Scoring").
			WithTextSize(18).
			WithPadding(5).
			WithOutlineWidth(2).
			WithRelativePos(vec.Vec2i{Y: 164}).
			WithTooltip(
				"Every 30s: Farm +2, Forester +2, Mine +3, Gold Mine +5, Barracks +3, Bank +3\n" +
					"Destroy enemies for points; eliminating every rival wins immediately",
			),
	)
	return group
}

var serverResourceDisplayOrder = []game.ResourceType{
	game.ResourceFood,
	game.ResourceWood,
	game.ResourceStone,
	game.ResourceIron,
	game.ResourceGold,
}

func NewGameScreen(previousScreen *ui.ScreenElement) *ui.ScreenElement {
	gamePreviousScreen = previousScreen
	// nil previousScreen → Leave Game creates a fresh main screen.
	escScreen = NewEscScreen(nil)

	var copiedGameCode string
	var gameCodeCopiedUntil time.Time

	gameCodeButtonText := func() string {
		if currentGame == nil || !currentGame.Multiplayer || currentGame.GameCode == "" {
			return ""
		}
		if currentGame.GameCode == copiedGameCode && time.Now().Before(gameCodeCopiedUntil) {
			return "Copied!"
		}
		return "Game code: " + currentGame.GameCode
	}

	transparent := color.RGBA{}
	gameCodeButton := ui.Button().
		WithTextDynamic(gameCodeButtonText).
		WithTextSize(28).
		WithPadding(0).
		WithOutlineWidth(0).
		WithoutShadow().
		WithForegroundColors(ui.ColorSet{
			Default: &rl.White,
			Hover:   &ui.PaletteIndigo,
			Click:   &ui.PaletteIndigoPress,
		}).
		WithBackgroundColors(ui.ColorSet{Default: &transparent}).
		WithOutlineColors(ui.ColorSet{Default: &transparent}).
		WithAnchors(anchor.TopRight, anchor.TopRight).
		WithRelativePos(vec.Vec2i{X: -20, Y: 20}).
		WithVisibleDynamic(func(el *ui.ButtonElement) bool {
			return currentGame != nil && currentGame.Multiplayer && currentGame.GameCode != ""
		}).
		WithClick(func() {
			if currentGame == nil || !currentGame.Multiplayer || currentGame.GameCode == "" {
				return
			}

			copiedGameCode = currentGame.GameCode
			gameCodeCopiedUntil = time.Now().Add(gameCodeCopyFeedbackDuration)
			rl.SetClipboardText(copiedGameCode)
		})

	screen := ui.Screen().
		WithEnter(func() {
			HideEscScreen()
			gameWorld.Renderer.ResetCamera(&gameWorld.Map)

			audio.StartMusic()
			audio.StartAmbience()
		}).
		WithExit(func() {
			HideEscScreen()
			gameWorld.Renderer.ResetCamera(&gameWorld.Map)
			audio.StopMusic()
			audio.StopAmbience()
			clearCurrentGame()
		}).
		// Screen children update back-to-front. Adding this before gameWorld
		// makes the world initialize first, then applies any deferred focus.
		AddChild(
			ui.Group().WithUpdate(func(deltaNano int64) {
				if focusTownhallPending && tryFocusOnTownhall() {
					focusTownhallPending = false
				}
			}),
		).
		AddChild(
			gameWorld,
		).
		AddChild(gameCodeButton).
		AddChild(
			ui.Text().
				WithTextDynamic(multiplayerStatusText).
				WithTextSize(26).
				WithTextColor(rl.White).
				WithAnchors(anchor.Top, anchor.Top).
				WithRelativePos(vec.Vec2i{X: 0, Y: 20}).
				WithVisibleDynamic(func(el *ui.TextElement) bool {
					return multiplayerStatusText() != ""
				}),
		).
		AddChild(
			ui.Button().
				WithText("Start Game").
				WithTextSize(30).
				WithPadding(10).
				WithOutlineWidth(4).
				WithAnchors(anchor.Top, anchor.Top).
				WithRelativePos(vec.Vec2i{X: 0, Y: 58}).
				WithVisibleDynamic(func(el *ui.ButtonElement) bool {
					return currentGame != nil && currentGame.Multiplayer &&
						!serverGameActive && gameOverMessage == "" && isLocalHost()
				}).
				WithClick(func() {
					gameActionError = ""
					if err := gameNet.Send(&packets.C2SStartGamePacket{}); err != nil {
						gameActionError = capitalizeSentence(err.Error())
						fmt.Printf("failed to request game start: %v\n", err)
					}
				}),
		).
		AddChild(
			ui.Text().
				WithTextDynamic(func() string { return gameActionError }).
				WithTextSize(22).
				WithTextColor(color.RGBA{R: 255, G: 96, B: 96, A: 255}).
				WithTextShadow(color.RGBA{R: 0, G: 0, B: 0, A: 200}, vec.Vec2i{X: 2, Y: 2}).
				WithAnchors(anchor.Top, anchor.Top).
				WithRelativePos(vec.Vec2i{X: 0, Y: 122}).
				WithVisibleDynamic(func(el *ui.TextElement) bool {
					return gameActionError != ""
				}),
		).
		AddChild(
			ui.Text().
				WithTextDynamic(func() string {
					if gameResolutionMessage == "" || time.Now().After(gameResolutionUntil) {
						return ""
					}
					return gameResolutionMessage
				}).
				WithTextSize(22).
				WithTextColor(rl.White).
				WithTextShadow(color.RGBA{A: 210}, vec.Vec2i{X: 2, Y: 2}).
				WithAnchors(anchor.Top, anchor.Top).
				WithRelativePos(vec.Vec2i{X: 0, Y: 154}).
				WithVisibleDynamic(func(el *ui.TextElement) bool {
					return gameResolutionMessage != "" && time.Now().Before(gameResolutionUntil)
				}),
		).
		AddChild(
			ui.Text().
				WithTextDynamic(func() string {
					if gamePendingAction == "" {
						return ""
					}
					return "Pending: " + gamePendingAction
				}).
				WithTextSize(20).
				WithTextColor(rl.White).
				WithAnchors(anchor.Bottom, anchor.Bottom).
				WithRelativePos(vec.Vec2i{X: 0, Y: -18}).
				WithVisibleDynamic(func(el *ui.TextElement) bool {
					return serverGameActive && gamePendingAction != ""
				}),
		).
		AddChild(buildingToolbar()).
		AddChild(
			ui.Group().
				WithAnchors(anchor.TopLeft, anchor.TopLeft).
				WithRelativePos(vec.Vec2i{X: 8, Y: 8}).
				WithVisibleDynamic(func(el *ui.GroupElement) bool {
					return !serverGameActive && global.DebugEnabled
				}).
				AddChild(
					gameSeedInput.
						WithPadding(8).
						WithTextSize(24).
						WithSize(vec.Vec2i{X: 320, Y: 0}).
						WithPlaceholderText("Seed").
						WithDefaultText(""),
				).
				AddChild(
					gameRegenerateButton.
						WithPadding(8).
						WithTextSize(24).
						WithRelativePos(vec.Vec2i{X: 0, Y: 52}).
						WithText("Regenerate").
						WithClick(func() {
							gameWorld.Map.Seed = gameSeedFromText(gameSeedInput.Value())
							gameWorld.Map.Generate()
						}),
				).
				AddChild(
					ui.Button().
						WithPadding(8).
						WithTextSize(24).
						WithRelativePos(vec.Vec2i{X: 0, Y: 104}).
						WithText("Random").
						WithClick(func() {
							gameSeedInput.SetText(strconv.FormatInt(rand.Int63(), 10))
							gameRegenerateButton.Click()
						}),
				),
		).
		AddChild(serverResourceList()).
		AddChild(
			ui.Text().
				WithTextDynamic(func() string {
					if serverGameEndTime <= 0 || !serverGameActive {
						return ""
					}
					gameEnd := max(time.Until(time.Unix(0, serverGameEndTime)), 0)
					if gameEnd >= time.Minute {
						return fmt.Sprintf("Ends in: %dm %ds",
							int(gameEnd.Minutes()), int(gameEnd.Seconds())%60)
					}
					return fmt.Sprintf("Ends in: %ds", int(gameEnd.Seconds()))
				}).
				WithTextSize(20).
				WithTextColor(rl.White).
				WithAnchors(anchor.TopLeft, anchor.TopLeft).
				WithRelativePos(vec.Vec2i{X: 8, Y: 8}),
		).
		AddChild(
			ui.Button().
				WithText("Town Hall").
				WithTextSize(18).
				WithPadding(6).
				WithOutlineWidth(2).
				WithAnchors(anchor.BottomRight, anchor.BottomRight).
				WithRelativePos(vec.Vec2i{X: -20, Y: -20}).
				WithVisibleDynamic(func(el *ui.ButtonElement) bool {
					return serverGameActive
				}).
				WithClick(focusOnTownhall),
		).
		AddChild(
			ui.Button().
				WithText("Hold Position").
				WithTextSize(18).
				WithPadding(6).
				WithOutlineWidth(2).
				WithAnchors(anchor.BottomRight, anchor.BottomRight).
				WithRelativePos(vec.Vec2i{X: -128, Y: -20}).
				WithVisibleDynamic(func(el *ui.ButtonElement) bool {
					return serverGameActive
				}).
				WithClick(func() {
					if err := gameNet.SendPassAction(serverRound); err != nil {
						fmt.Printf("failed to hold round: %v\n", err)
						return
					}
					gameWorld.Renderer.ClearQueuedBuilding()
					setPendingAction("Hold position")
				}),
		).
		AddChild(
			ui.GameBuildingDetailsPanel().
				WithWorld(gameWorld),
		).
		AddChild(
			ui.Vignette().WithAlpha(120),
		).
		// The real round countdown stays below the pause screen, keeping the
		// settings controls unobstructed when both are visible.
		AddChild(
			ui.GameSelectionMenu().
				WithWorld(gameWorld),
		).
		AddChild(
			newRoundCountdown(),
		).
		AddChild(
			escScreen,
		)

	gameScreen = screen
	return screen
}
