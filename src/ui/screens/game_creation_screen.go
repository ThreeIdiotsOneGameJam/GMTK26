package screens

import (
	"fmt"
	"strconv"

	gameNet "github.com/threeidiotsonegamejam/gmtk26/src/net"
	"github.com/threeidiotsonegamejam/gmtk26/src/net/packets"
	"github.com/threeidiotsonegamejam/gmtk26/src/settings"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui/anchor"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui/uiutil"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

type gameCreationMode uint8

const (
	gameCreationSolo gameCreationMode = iota
	gameCreationHost
)

var (
	creationMode        = gameCreationSolo
	creationMaxPlayers  = uint8(4)
	creationPublic      = true
	creationSubmitting  bool
	creationError       string
	creationSubmit      *ui.ButtonElement
	creationVisibility  *ui.ButtonElement
	creationPlayerCount [3]*ui.ButtonElement
)

func OpenSoloGameCreation(previousScreen *ui.ScreenElement) {
	openGameCreation(gameCreationSolo, previousScreen)
}

func OpenHostGameCreation(previousScreen *ui.ScreenElement) {
	openGameCreation(gameCreationHost, previousScreen)
}

func RejectGameCreation(message string) {
	if gameNet.LocalGameActive() {
		gameNet.StopLocalGame()
	}
	creationSubmitting = false
	creationError = message
	SetPlayError(message)
}

func StartSoloWithDefaults() error {
	return startSoloGame()
}

func HostGameWithDefaults() error {
	return sendHostGame(true, 4)
}

func openGameCreation(mode gameCreationMode, previousScreen *ui.ScreenElement) {
	creationMode = mode
	creationMaxPlayers = 4
	creationPublic = true
	creationSubmitting = false
	creationError = ""
	screen := NewGameCreationScreen(previousScreen)
	SetActiveScreen(screen)
}

func NewGameCreationScreen(previousScreen *ui.ScreenElement) *ui.ScreenElement {
	hostMode := func() bool {
		return creationMode == gameCreationHost
	}
	hostConnected := func() bool {
		return !settings.Current.Offline && gameNet.State() == gameNet.ConnectionConnected
	}

	controller := ui.Group().WithUpdate(func(deltaNano int64) {
		if creationSubmit != nil {
			switch {
			case creationSubmitting:
				creationSubmit.Text = "Creating..."
			case hostMode():
				creationSubmit.Text = "Create Game"
			default:
				creationSubmit.Text = "Start Solo Game"
			}
		}
		if creationVisibility != nil {
			if creationPublic {
				creationVisibility.Text = "Visibility: Public"
			} else {
				creationVisibility.Text = "Visibility: Code Only"
			}
		}
		for i, button := range creationPlayerCount {
			count := uint8(i + 2)
			if count == creationMaxPlayers {
				button.Text = fmt.Sprintf("[%d]", count)
			} else {
				button.Text = strconv.Itoa(int(count))
			}
		}

		if creationSubmitting && hostMode() && !hostConnected() {
			creationSubmitting = false
			creationError = "Connection lost before the game was created"
		}
	})

	for i := range creationPlayerCount {
		count := uint8(i + 2)
		creationPlayerCount[i] = ui.Button().
			WithText(strconv.Itoa(int(count))).
			WithTextSize(30).
			WithPadding(8).
			WithSize(vec.Vec2i{X: 100, Y: 52}).
			WithAnchors(anchor.Center, anchor.Center).
			WithRelativePos(vec.Vec2i{X: int32(i-1) * 120, Y: -22}).
			WithVisibleDynamic(func(el *ui.ButtonElement) bool {
				return hostMode()
			}).
			WithEnabledDynamic(func(el *ui.ButtonElement) bool {
				return !creationSubmitting
			}).
			WithClick(func() {
				creationMaxPlayers = count
				creationError = ""
			})
	}

	creationVisibility = ui.Button().
		WithText("Visibility: Public").
		WithTextSize(28).
		WithPadding(8).
		WithSize(vec.Vec2i{X: 360, Y: 52}).
		WithAnchors(anchor.Center, anchor.Center).
		WithRelativePos(vec.Vec2i{X: 0, Y: 48}).
		WithVisibleDynamic(func(el *ui.ButtonElement) bool {
			return hostMode()
		}).
		WithEnabledDynamic(func(el *ui.ButtonElement) bool {
			return !creationSubmitting
		}).
		WithClick(func() {
			creationPublic = !creationPublic
			creationError = ""
		})

	creationSubmit = ui.Button().
		WithText("Start Solo Game").
		WithTextSize(34).
		WithPadding(10).
		WithOutlineWidth(4).
		WithSize(vec.Vec2i{X: 360, Y: 58}).
		WithAnchors(anchor.Center, anchor.Center).
		WithRelativePosDynamic(func(el *ui.ButtonElement) vec.Vec2i {
			if hostMode() {
				return vec.Vec2i{X: 0, Y: 114}
			}
			return vec.Vec2i{X: 0, Y: 48}
		}).
		WithEnabledDynamic(func(el *ui.ButtonElement) bool {
			return !creationSubmitting && (!hostMode() || hostConnected())
		}).
		WithClick(submitGameCreation)

	screen := ui.Screen().
		WithBackgroundColor(uiutil.MenuScreenBackground).
		AddChild(uiutil.MenuBackdrop()).
		AddChild(controller).
		AddChild(
			ui.Text().
				WithTextDynamic(func() string {
					if hostMode() {
						return "Host a Game"
					}
					return "Create Solo Game"
				}).
				WithTextSize(76).
				WithTextColor(uiutil.MenuHeaderColor).
				WithAnchors(anchor.Center, anchor.Top).
				WithRelativePos(vec.Vec2i{X: 0, Y: 42}),
		).
		AddChild(
			ui.Text().
				WithTextDynamic(func() string {
					if hostMode() {
						return "Choose the rules"
					}
					return "The world is generated when the game starts"
				}).
				WithTextSize(26).
				WithTextColor(uiutil.MenuMutedColor).
				WithAnchors(anchor.Center, anchor.Top).
				WithRelativePos(vec.Vec2i{X: 0, Y: 122}),
		).
		AddChild(
			ui.Text().
				WithText("Maximum Players").
				WithTextSize(25).
				WithTextColor(uiutil.MenuMutedColor).
				WithAnchors(anchor.Center, anchor.Center).
				WithRelativePos(vec.Vec2i{X: 0, Y: -68}).
				WithVisibleDynamic(func(el *ui.TextElement) bool {
					return hostMode()
				}),
		).
		AddChild(
			ui.Text().
				WithText("Play against three AI factions").
				WithTextSize(26).
				WithTextColor(uiutil.MenuMutedColor).
				WithAnchors(anchor.Center, anchor.Center).
				WithRelativePos(vec.Vec2i{X: 0, Y: -18}).
				WithVisibleDynamic(func(el *ui.TextElement) bool {
					return !hostMode()
				}),
		)

	for _, button := range creationPlayerCount {
		screen.AddChild(button)
	}

	goBack := func() {
		if creationSubmitting {
			return
		}
		GoToPreviousScreen(previousScreen)
	}

	return screen.
		AddChild(creationVisibility).
		AddChild(creationSubmit).
		AddChild(
			ui.Text().
				WithTextDynamic(gameCreationStatus).
				WithTextSize(22).
				WithTextColor(ui.PaletteNegative).
				WithAnchors(anchor.Bottom, anchor.Bottom).
				WithRelativePos(vec.Vec2i{X: 0, Y: -44}).
				WithVisibleDynamic(func(el *ui.TextElement) bool {
					return gameCreationStatus() != ""
				}),
		).
		AddChild(
			backButton(goBack).WithEnabledDynamic(func(el *ui.ButtonElement) bool {
				return !creationSubmitting
			}),
		).
		AddChild(ui.Vignette()).
		WithBack(goBack)
}

func submitGameCreation() {
	creationError = ""

	if creationMode == gameCreationSolo {
		if err := startSoloGame(); err != nil {
			creationError = err.Error()
			return
		}
		creationSubmitting = true
		return
	}

	if err := sendHostGame(creationPublic, creationMaxPlayers); err != nil {
		creationError = err.Error()
		return
	}
	creationSubmitting = true
}

func startSoloGame() error {
	return gameNet.StartLocalGame()
}

func sendHostGame(public bool, maxPlayers uint8) error {
	if settings.Current.Offline || gameNet.State() != gameNet.ConnectionConnected {
		return fmt.Errorf("Connect to the multiplayer server before creating a game")
	}
	if err := gameNet.Send(&packets.C2SCreateGamePacket{
		Public:     public,
		MaxPlayers: maxPlayers,
	}); err != nil {
		return err
	}
	return nil
}

func gameCreationStatus() string {
	if creationError != "" {
		return creationError
	}
	if creationMode == gameCreationHost && (settings.Current.Offline || gameNet.State() != gameNet.ConnectionConnected) {
		return "Multiplayer connection required"
	}
	return ""
}
