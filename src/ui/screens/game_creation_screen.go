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
	creationMode       = gameCreationSolo
	creationMaxPlayers = uint8(4)
	creationPublic     = true
	creationSubmitting bool
	creationError      string
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
	message = capitalizeSentence(message)
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
		if creationSubmitting && hostMode() && !hostConnected() {
			creationSubmitting = false
			creationError = "Connection lost before the game was created"
		}
	})

	playerCountButtons := make([]ui.Element, 0, 3)
	for i := range 3 {
		count := uint8(i + 2)
		playerCountButtons = append(playerCountButtons, ui.Button().
			WithTextDynamic(func() string {
				if count == creationMaxPlayers {
					return fmt.Sprintf("[%d]", count)
				}
				return strconv.Itoa(int(count))
			}).
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
			}))
	}

	creationVisibility := ui.Button().
		WithTextDynamic(func() string {
			if creationPublic {
				return "Visibility: Public"
			}
			return "Visibility: Code Only"
		}).
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

	creationSubmit := ui.Button().
		WithTextDynamic(func() string {
			switch {
			case creationSubmitting:
				return "Creating..."
			case hostMode():
				return "Create Game"
			default:
				return "Start Solo Game"
			}
		}).
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

	screen := uiutil.MenuScreen().
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

	screen.AddChildren(playerCountButtons...)

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
			uiutil.BackButton(goBack).WithEnabledDynamic(func(el *ui.ButtonElement) bool {
				return !creationSubmitting
			}),
		).
		AddChild(uiutil.MenuVignette()).
		WithBack(goBack)
}

func submitGameCreation() {
	creationError = ""

	if creationMode == gameCreationSolo {
		if err := startSoloGame(); err != nil {
			creationError = capitalizeSentence(err.Error())
			return
		}
		creationSubmitting = true
		return
	}

	if err := sendHostGame(creationPublic, creationMaxPlayers); err != nil {
		creationError = capitalizeSentence(err.Error())
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
