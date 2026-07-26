package screens

import (
	"fmt"
	"hash/fnv"
	"math/rand"
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
	creationSeedInput  *ui.InputElement
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
	return startSoloGame(rand.Int63())
}

func HostGameWithDefaults() error {
	return sendHostGame(true, 4, rand.Int63())
}

func openGameCreation(mode gameCreationMode, previousScreen *ui.ScreenElement) {
	creationMode = mode
	creationMaxPlayers = 4
	creationPublic = true
	creationSubmitting = false
	creationError = ""
	screen := NewGameCreationScreen(previousScreen)
	creationSeedInput.SetText(strconv.FormatInt(rand.Int63(), 10))
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

	creationSeedInput = ui.Input().
		WithPlaceholderText("Seed or phrase").
		WithMaxTextLength(48).
		WithDefaultText("").
		WithTextSize(30).
		WithPadding(10).
		WithSize(vec.Vec2i{X: 360, Y: 54}).
		WithAnchors(anchor.Center, anchor.Center).
		WithRelativePos(vec.Vec2i{X: -82, Y: -72}).
		WithEnabledDynamic(func(el *ui.InputElement) bool {
			return !creationSubmitting
		}).
		WithCallback(func(text string) {
			creationError = ""
		})

	randomSeedButton := ui.Button().
		WithText("Random").
		WithTextSize(28).
		WithPadding(8).
		WithSize(vec.Vec2i{X: 150, Y: 54}).
		WithAnchors(anchor.Center, anchor.Center).
		WithRelativePos(vec.Vec2i{X: 205, Y: -72}).
		WithEnabledDynamic(func(el *ui.ButtonElement) bool {
			return !creationSubmitting
		}).
		WithClick(func() {
			creationSeedInput.SetText(strconv.FormatInt(rand.Int63(), 10))
			creationError = ""
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
			WithRelativePos(vec.Vec2i{X: int32(i-1) * 120, Y: 34}).
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
		WithRelativePos(vec.Vec2i{X: 0, Y: 104}).
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
				return vec.Vec2i{X: 0, Y: 170}
			}
			return vec.Vec2i{X: 0, Y: 82}
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
					return "Choose a world seed before starting"
				}).
				WithTextSize(26).
				WithTextColor(uiutil.MenuMutedColor).
				WithAnchors(anchor.Center, anchor.Top).
				WithRelativePos(vec.Vec2i{X: 0, Y: 122}),
		).
		AddChild(
			ui.Text().
				WithText("World Seed").
				WithTextSize(25).
				WithTextColor(uiutil.MenuMutedColor).
				WithAnchors(anchor.Center, anchor.Center).
				WithRelativePos(vec.Vec2i{X: 0, Y: -116}),
		).
		AddChild(creationSeedInput).
		AddChild(randomSeedButton).
		AddChild(
			ui.Text().
				WithText("Maximum Players").
				WithTextSize(25).
				WithTextColor(uiutil.MenuMutedColor).
				WithAnchors(anchor.Center, anchor.Center).
				WithRelativePos(vec.Vec2i{X: 0, Y: -12}).
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
				WithRelativePos(vec.Vec2i{X: 0, Y: 10}).
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
		WithBack(goBack).
		WithExit(func() {
			creationSeedInput.Blur()
		})
}

func submitGameCreation() {
	seed := gameSeedFromText(creationSeedInput.Value())
	creationError = ""

	if creationMode == gameCreationSolo {
		if err := startSoloGame(seed); err != nil {
			creationError = capitalizeSentence(err.Error())
			return
		}
		creationSubmitting = true
		return
	}

	if err := sendHostGame(creationPublic, creationMaxPlayers, seed); err != nil {
		creationError = capitalizeSentence(err.Error())
		return
	}
	creationSubmitting = true
}

func startSoloGame(seed int64) error {
	return gameNet.StartLocalGame(seed)
}

func sendHostGame(public bool, maxPlayers uint8, seed int64) error {
	if settings.Current.Offline || gameNet.State() != gameNet.ConnectionConnected {
		return fmt.Errorf("Connect to the multiplayer server before creating a game")
	}
	if err := gameNet.Send(&packets.C2SCreateGamePacket{
		Public:     public,
		MaxPlayers: maxPlayers,
		Seed:       seed,
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

func gameSeedFromText(text string) int64 {
	if text == "" || text == "0" {
		return 0
	}
	if seed, err := strconv.ParseInt(text, 10, 64); err == nil {
		return seed
	}

	hash := fnv.New64a()
	_, _ = hash.Write([]byte(text))
	return int64(hash.Sum64())
}
