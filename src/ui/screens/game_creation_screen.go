package screens

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"strconv"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	gameNet "github.com/threeidiotsonegamejam/gmtk26/src/net"
	"github.com/threeidiotsonegamejam/gmtk26/src/net/packets"
	"github.com/threeidiotsonegamejam/gmtk26/src/settings"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui/anchor"
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
	creationSeedInput   *ui.InputElement
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
	creationSubmitting = false
	creationError = message
}

func openGameCreation(mode gameCreationMode, previousScreen *ui.ScreenElement) {
	creationMode = mode
	creationMaxPlayers = 4
	creationPublic = true
	creationSubmitting = false
	creationError = ""
	screen := NewGameCreationScreen(previousScreen)
	creationSeedInput.Text = strconv.FormatInt(rand.Int63(), 10)
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

	creationSeedInput = ui.Input().
		WithPlaceholderText("Seed or phrase").
		WithMaxTextLength(48).
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
			creationSeedInput.Text = strconv.FormatInt(rand.Int63(), 10)
			creationError = ""
		})

	for i := range creationPlayerCount {
		count := uint8(i + 2)
		creationPlayerCount[i] = ui.Button().
			WithText(strconv.Itoa(int(count))).
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
			})
	}

	creationVisibility = ui.Button().
		WithText("Visibility: Public").
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

	creationSubmit = ui.Button().
		WithText("Start Solo Game").
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

	screen := ui.Screen().
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
				WithTextColor(rl.Black).
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
				WithTextColor(rl.Gray).
				WithAnchors(anchor.Center, anchor.Top).
				WithRelativePos(vec.Vec2i{X: 0, Y: 122}),
		).
		AddChild(
			ui.Text().
				WithText("World Seed").
				WithTextSize(25).
				WithTextColor(rl.DarkGray).
				WithAnchors(anchor.Center, anchor.Center).
				WithRelativePos(vec.Vec2i{X: 0, Y: -116}),
		).
		AddChild(creationSeedInput).
		AddChild(randomSeedButton).
		AddChild(
			ui.Text().
				WithText("Maximum Players").
				WithTextSize(25).
				WithTextColor(rl.DarkGray).
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
				WithTextColor(rl.DarkGray).
				WithAnchors(anchor.Center, anchor.Center).
				WithRelativePos(vec.Vec2i{X: 0, Y: 10}).
				WithVisibleDynamic(func(el *ui.TextElement) bool {
					return !hostMode()
				}),
		)

	for _, button := range creationPlayerCount {
		screen.AddChild(button)
	}

	return screen.
		AddChild(creationVisibility).
		AddChild(creationSubmit).
		AddChild(
			ui.Text().
				WithTextDynamic(gameCreationStatus).
				WithTextSize(22).
				WithTextColor(rl.Maroon).
				WithAnchors(anchor.Bottom, anchor.Bottom).
				WithRelativePos(vec.Vec2i{X: 0, Y: -44}).
				WithVisibleDynamic(func(el *ui.TextElement) bool {
					return gameCreationStatus() != ""
				}),
		).
		AddChild(
			backButton(func() {
				GoToPreviousScreen(previousScreen)
			}).WithEnabledDynamic(func(el *ui.ButtonElement) bool {
				return !creationSubmitting
			}),
		).
		AddChild(ui.Vignette()).
		WithExit(func() {
			creationSeedInput.Blur()
		})
}

func submitGameCreation() {
	seed := gameSeedFromText(creationSeedInput.Text)
	creationError = ""

	if creationMode == gameCreationSolo {
		player := *game.PlayerData
		state := game.Game{
			HostID:      player.ClientID,
			Multiplayer: false,
			MaxPlayers:  1,
			Round:       1,
			Map: game.Map{
				Seed: seed,
			},
		}
		state.Factions[0].Player = &player
		for i := 1; i < len(state.Factions); i++ {
			state.Factions[i].AI = true
		}
		EnterGame(state)
		return
	}

	if settings.Current.Offline || gameNet.State() != gameNet.ConnectionConnected {
		creationError = "Connect to the multiplayer server before creating a game"
		return
	}
	if err := gameNet.Send(&packets.C2SCreateGamePacket{
		Public:     creationPublic,
		MaxPlayers: creationMaxPlayers,
		Seed:       seed,
	}); err != nil {
		creationError = err.Error()
		return
	}
	creationSubmitting = true
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
