package screens

import (
	"fmt"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	gameNet "github.com/threeidiotsonegamejam/gmtk26/src/net"
	"github.com/threeidiotsonegamejam/gmtk26/src/net/packets"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui/anchor"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

var (
	matchmakingActive   bool
	matchmakingQueuePos int
	matchmakingStarted  time.Time
	matchmakingPrevious *ui.ScreenElement
)

func EnterMatchmakingWaiting(previousScreen *ui.ScreenElement) {
	matchmakingActive = true
	matchmakingQueuePos = 0
	matchmakingStarted = time.Now()
	SetActiveScreen(NewMatchmakingScreen(previousScreen))
}

func ApplyMatchmakingWaiting(position int) {
	matchmakingActive = true
	matchmakingQueuePos = position
	if GetActiveScreen() == matchmakingScreen {
		return
	}
	previous := playScreen
	if previous == nil {
		previous = GetActiveScreen()
	}
	SetActiveScreen(NewMatchmakingScreen(previous))
}

func CancelMatchmaking() {
	if !matchmakingActive {
		return
	}
	if err := gameNet.Send(&packets.C2SLeaveGamePacket{}); err != nil {
		fmt.Printf("failed to cancel matchmaking: %v\n", err)
	}
	clearMatchmaking()
	GoToPreviousScreen(matchmakingPrevious)
}

func clearMatchmaking() {
	matchmakingActive = false
	matchmakingQueuePos = 0
}

func NewMatchmakingScreen(previousScreen *ui.ScreenElement) *ui.ScreenElement {
	matchmakingPrevious = previousScreen

	screen := ui.Screen().
		WithEnter(HideEscScreen).
		AddChild(
			ui.Group().WithUpdate(func(deltaNano int64) {
				if !matchmakingActive {
					return
				}
				if gameNet.State() != gameNet.ConnectionConnected {
					clearMatchmaking()
					SetPlayError("Connection lost while searching for a game")
					GoToPreviousScreen(previousScreen)
				}
			}),
		).
		AddChild(
			ui.Text().
				WithText("Finding a Game").
				WithTextSize(72).
				WithTextColor(rl.Black).
				WithAnchors(anchor.Center, anchor.Top).
				WithRelativePos(vec.Vec2i{X: 0, Y: 80}),
		).
		AddChild(
			ui.Text().
				WithTextDynamic(func() string {
					if matchmakingQueuePos > 0 {
						return fmt.Sprintf(
							"Waiting for a public lobby - position %d in queue",
							matchmakingQueuePos,
						)
					}
					return "Searching for a public lobby..."
				}).
				WithTextSize(28).
				WithTextColor(rl.DarkGray).
				WithAnchors(anchor.Center, anchor.Center).
				WithRelativePos(vec.Vec2i{X: 0, Y: -24}),
		).
		AddChild(
			ui.Text().
				WithTextDynamic(func() string {
					switch int(time.Since(matchmakingStarted).Seconds()) % 3 {
					case 0:
						return "Please wait."
					case 1:
						return "Please wait.."
					default:
						return "Please wait..."
					}
				}).
				WithTextSize(26).
				WithTextColor(rl.Gray).
				WithAnchors(anchor.Center, anchor.Center).
				WithRelativePos(vec.Vec2i{X: 0, Y: 24}),
		).
		AddChild(
			ui.Button().
				WithText("Cancel").
				WithTextSize(34).
				WithPadding(10).
				WithOutlineWidth(4).
				WithSize(vec.Vec2i{X: 240, Y: 56}).
				WithAnchors(anchor.Center, anchor.Center).
				WithRelativePos(vec.Vec2i{X: 0, Y: 120}).
				WithClick(CancelMatchmaking),
		).
		AddChild(ui.Vignette()).
		WithBack(CancelMatchmaking)

	matchmakingScreen = screen
	return screen
}
