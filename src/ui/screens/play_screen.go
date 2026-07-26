package screens

import (
	"fmt"
	"strings"
	"time"

	"github.com/threeidiotsonegamejam/gmtk26/src/game"
	gameNet "github.com/threeidiotsonegamejam/gmtk26/src/net"
	"github.com/threeidiotsonegamejam/gmtk26/src/net/packets"
	"github.com/threeidiotsonegamejam/gmtk26/src/settings"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui/anchor"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui/uiutil"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

const (
	gameCodeCharset          = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	gameCodeLength           = 6
	joinPanelTransition      = 220 * time.Millisecond
	playErrorDisplayDuration = 3 * time.Second
)

var (
	playGameCodeInput      *ui.InputElement
	playError              string
	playErrorTimeRemaining time.Duration
)

func ClearGameCodeInput() {
	if playGameCodeInput == nil {
		return
	}
	playGameCodeInput.SetText("")
	playGameCodeInput.Blur()
}

func SetPlayError(message string) {
	playError = capitalizeSentence(message)
	if playError == "" {
		playErrorTimeRemaining = 0
		return
	}
	playErrorTimeRemaining = playErrorDisplayDuration
}

func advancePlayError(delta time.Duration) {
	if playError == "" {
		return
	}

	playErrorTimeRemaining -= delta
	if playErrorTimeRemaining <= 0 {
		SetPlayError("")
	}
}

func NewPlayScreen(previousScreen *ui.ScreenElement) *ui.ScreenElement {
	var screen *ui.ScreenElement
	var joinPanelOpen bool
	var joinPanelProgress float32
	var focusCodeInput bool

	var codeInput *ui.InputElement

	multiplayerEnabled := func() bool {
		return !settings.Current.Offline && gameNet.State() == gameNet.ConnectionConnected
	}

	connectionWarning := func() string {
		if settings.Current.Offline {
			return "Multiplayer is disabled while offline mode is enabled"
		}
		if gameNet.State() == gameNet.ConnectionConnecting {
			return "Connecting to the multiplayer server..."
		}
		return "Multiplayer server unavailable. Retrying automatically..."
	}

	resetJoinPanel := func() {
		joinPanelOpen = false
		joinPanelProgress = 0
		focusCodeInput = false

		if codeInput != nil {
			codeInput.Blur()
		}
	}

	goBack := func() {
		if joinPanelOpen || joinPanelProgress > 0 {
			resetJoinPanel()
			return
		}
		GoToPreviousScreen(previousScreen)
	}

	sendGameRequest := func(packet packets.C2SPacket) {
		if err := gameNet.Send(packet); err != nil {
			SetPlayError(err.Error())
			fmt.Printf("failed to send game request: %v\n", err)
		}
	}

	joinGame := func(code string) {
		SetPlayError("")
		sendGameRequest(&packets.C2SJoinGamePacket{GameCode: code})
	}

	joinRandom := func() {
		SetPlayError("")
		EnterMatchmakingWaiting(screen)
		if err := gameNet.Send(&packets.C2SJoinGamePacket{GameCode: ""}); err != nil {
			clearMatchmaking()
			SetPlayError(err.Error())
			SetActiveScreen(screen)
			fmt.Printf("failed to start matchmaking: %v\n", err)
		}
	}

	panelProgress := func() float32 {
		return ui.Smoothstep(joinPanelProgress)
	}

	joinPanelMessagePos := func(el *ui.TextElement) vec.Vec2i {
		if joinPanelProgress > 0 {
			progress := panelProgress()
			return vec.Vec2i{Y: 244 + int32((1-progress)*16)}
		}

		return vec.Vec2i{
			Y: el.Parent.Size().Y/2 - 104 - el.Size().Y/2,
		}
	}

	menuButtonPos := func(openY int32) func(*ui.ButtonElement) vec.Vec2i {
		return func(el *ui.ButtonElement) vec.Vec2i {
			closedOffset := el.Parent.Size().Y / 12
			progress := panelProgress()

			return vec.Vec2i{
				X: 0,
				Y: openY + int32(float32(closedOffset)*(1-progress)),
			}
		}
	}

	controller := ui.Group().WithUpdate(func(deltaNano int64) {
		advancePlayError(time.Duration(deltaNano))

		if !multiplayerEnabled() && joinPanelOpen {
			joinPanelOpen = false
			focusCodeInput = false
			codeInput.Blur()
		}

		target := float32(0)
		if joinPanelOpen {
			target = 1
		}

		joinPanelProgress = ui.MoveTowards(
			joinPanelProgress,
			target,
			float32(time.Duration(deltaNano))/float32(joinPanelTransition),
		)

		if focusCodeInput && joinPanelProgress >= 0.98 && codeInput != nil {
			codeInput.Focus()
			focusCodeInput = false
		}
	})

	codeIsValid := func() bool {
		return codeInput != nil && len([]rune(codeInput.Value())) == gameCodeLength
	}

	codeInput = ui.Input().
		WithPlaceholderText("GAME CODE").
		WithMaxTextLength(gameCodeLength).
		WithCharset(gameCodeCharset).
		WithInputTransformer(strings.ToUpper).
		WithDefaultText("").
		WithTextSize(34).
		WithPadding(10).
		WithSize(vec.Vec2i{X: 280, Y: 58}).
		WithAnchors(anchor.Center, anchor.Center).
		WithRelativePosDynamic(func(el *ui.InputElement) vec.Vec2i {
			progress := panelProgress()
			return vec.Vec2i{
				X: -80,
				Y: 190 + int32((1-progress)*16),
			}
		}).
		WithVisibleDynamic(func(el *ui.InputElement) bool {
			return joinPanelProgress > 0
		}).
		WithEnabledDynamic(func(el *ui.InputElement) bool {
			return multiplayerEnabled() && joinPanelOpen && joinPanelProgress >= 0.98
		}).
		WithOpacityDynamic(func(el *ui.InputElement) float32 {
			return panelProgress()
		}).
		WithSubmit(func(text string) {
			if codeIsValid() {
				joinGame(text)
			}
		})
	playGameCodeInput = codeInput

	joinButton := playMenuButton("Join").
		WithTextSize(34).
		WithSize(vec.Vec2i{X: 144, Y: 58}).
		WithRelativePosDynamic(func(el *ui.ButtonElement) vec.Vec2i {
			progress := panelProgress()
			return vec.Vec2i{
				X: 148,
				Y: 190 + int32((1-progress)*16),
			}
		}).
		WithVisibleDynamic(func(el *ui.ButtonElement) bool {
			return joinPanelProgress > 0
		}).
		WithEnabledDynamic(func(el *ui.ButtonElement) bool {
			return multiplayerEnabled() && joinPanelOpen && joinPanelProgress >= 0.98 && codeIsValid()
		}).
		WithOpacityDynamic(func(el *ui.ButtonElement) float32 {
			opacity := panelProgress()
			if !codeIsValid() {
				opacity *= 0.45
			}
			return opacity
		}).
		WithClick(func() {
			joinGame(codeInput.Value())
		})

	codeHelp := ui.Text().
		WithTextDynamic(func() string {
			if codeIsValid() {
				return "Press Enter or Join"
			}
			return "Enter a 6-character game code"
		}).
		WithTextSize(24).
		WithTextColor(uiutil.MenuMutedColor).
		WithAnchors(anchor.Center, anchor.Center).
		WithRelativePosDynamic(joinPanelMessagePos).
		WithVisibleDynamic(func(el *ui.TextElement) bool {
			return joinPanelProgress > 0 && playError == ""
		}).
		WithOpacityDynamic(func(el *ui.TextElement) float32 {
			return panelProgress()
		})

	joinCodeButton := playMenuButton("Join with Game Code").
		WithTextDynamic(func() string {
			if joinPanelOpen {
				return "Cancel Game Code"
			}
			return "Join with Game Code"
		}).
		WithRelativePosDynamic(menuButtonPos(116)).
		WithEnabledDynamic(func(el *ui.ButtonElement) bool {
			return multiplayerEnabled()
		}).
		WithClick(func() {
			joinPanelOpen = !joinPanelOpen
			if joinPanelOpen {
				focusCodeInput = true
			} else {
				codeInput.Blur()
			}
		})

	screen = uiutil.MenuScreen().
		AddChild(controller).
		AddChild(
			ui.Text().
				WithText("Play").
				WithTextSize(88).
				WithTextColor(uiutil.MenuHeaderColor).
				WithAnchors(anchor.Center, anchor.Top).
				WithRelativePosDynamic(func(el *ui.TextElement) vec.Vec2i {
					return vec.Vec2i{X: 0, Y: max(int32(42), el.Parent.Size().Y/8)}
				}),
		).
		AddChild(
			ui.Text().
				WithText("Choose how to play").
				WithTextSize(28).
				WithTextColor(uiutil.MenuMutedColor).
				WithAnchors(anchor.Center, anchor.Top).
				WithRelativePosDynamic(func(el *ui.TextElement) vec.Vec2i {
					return vec.Vec2i{X: 0, Y: max(int32(136), el.Parent.Size().Y/8+92)}
				}),
		).
		AddChild(
			ui.Input().
				WithDefaultText(game.PlayerData.PlayerName).
				WithPlaceholderText("Your Name").
				WithMaxTextLength(20).
				WithTextSize(28).
				WithPadding(8).
				WithSize(vec.Vec2i{X: 440, Y: 48}).
				WithAnchors(anchor.Center, anchor.Center).
				WithRelativePosDynamic(func(el *ui.InputElement) vec.Vec2i {
					closedOffset := el.Parent.Size().Y / 12
					progress := panelProgress()
					return vec.Vec2i{
						Y: -172 + int32(float32(closedOffset)*(1-progress)),
					}
				}).
				WithCallback(func(text string) {
					game.PlayerData.PlayerName = text
					game.SavePlayerData()
				}),
		).
		AddChild(
			playMenuButton("Play Solo").
				WithRelativePosDynamic(menuButtonPos(-100)).
				WithClick(func() {
					if ui.DebugQuickActionModifierHeld() {
						SetPlayError("")
						if err := StartSoloWithDefaults(); err != nil {
							SetPlayError(err.Error())
						}
						return
					}
					OpenSoloGameCreation(screen)
				}),
		).
		AddChild(
			playMenuButton("Host a Game").
				WithRelativePosDynamic(menuButtonPos(-28)).
				WithEnabledDynamic(func(el *ui.ButtonElement) bool {
					return multiplayerEnabled()
				}).
				WithClick(func() {
					if ui.DebugQuickActionModifierHeld() {
						SetPlayError("")
						if err := HostGameWithDefaults(); err != nil {
							SetPlayError(err.Error())
						}
						return
					}
					OpenHostGameCreation(screen)
				}),
		).
		AddChild(
			playMenuButton("Join Random Game").
				WithRelativePosDynamic(menuButtonPos(44)).
				WithEnabledDynamic(func(el *ui.ButtonElement) bool {
					return multiplayerEnabled()
				}).
				WithClick(joinRandom),
		).
		AddChild(joinCodeButton).
		AddChild(codeInput).
		AddChild(joinButton).
		AddChild(codeHelp).
		AddChild(
			ui.Text().
				WithTextDynamic(func() string { return playError }).
				WithTextSize(22).
				WithTextColor(ui.PaletteNegative).
				WithAnchors(anchor.Center, anchor.Center).
				WithRelativePosDynamic(joinPanelMessagePos).
				WithVisibleDynamic(func(el *ui.TextElement) bool {
					return playError != ""
				}),
		).
		AddChild(
			ui.Text().
				WithTextDynamic(connectionWarning).
				WithTextSize(22).
				WithTextColor(uiutil.MenuMutedColor).
				WithAnchors(anchor.Bottom, anchor.Bottom).
				WithRelativePos(vec.Vec2i{X: 0, Y: -80}).
				WithVisibleDynamic(func(el *ui.TextElement) bool {
					return playError == "" && !multiplayerEnabled() && joinPanelProgress <= 0
				}),
		).
		AddChild(
			playMenuButton("Try Reconnect").
				WithTextSize(26).
				WithSize(vec.Vec2i{X: 240, Y: 48}).
				WithAnchors(anchor.Bottom, anchor.Bottom).
				WithRelativePos(vec.Vec2i{X: 0, Y: -24}).
				WithVisibleDynamic(func(el *ui.ButtonElement) bool {
					return !settings.Current.Offline && gameNet.State() == gameNet.ConnectionDisconnected && joinPanelProgress <= 0
				}).
				WithClick(gameNet.RetryConnection),
		).
		AddChild(uiutil.BackButton(goBack)).
		AddChild(uiutil.MenuVignette()).
		WithBack(goBack).
		WithExit(resetJoinPanel)

	playScreen = screen
	return screen
}

func playMenuButton(text string) *ui.ButtonElement {
	return ui.Button().
		WithText(text).
		WithTextSize(36).
		WithPadding(10).
		WithOutlineWidth(4).
		WithSize(vec.Vec2i{X: 440, Y: 58}).
		WithAnchors(anchor.Center, anchor.Center)
}
