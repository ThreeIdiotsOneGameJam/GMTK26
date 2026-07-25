package screens

import (
	"fmt"
	"strings"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	gameNet "github.com/threeidiotsonegamejam/gmtk26/src/net"
	"github.com/threeidiotsonegamejam/gmtk26/src/net/packets"
	"github.com/threeidiotsonegamejam/gmtk26/src/settings"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui/anchor"
	"github.com/threeidiotsonegamejam/gmtk26/src/util"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

const (
	gameCodeCharset     = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	gameCodeLength      = 6
	joinPanelTransition = 220 * time.Millisecond
)

var PlayScreen = newPlayScreen()

func newPlayScreen() *ui.ScreenElement {
	var joinPanelOpen bool
	var joinPanelProgress float32
	var focusCodeInput bool

	var codeInput *ui.InputElement
	var joinCodeButton *ui.ButtonElement

	multiplayerEnabled := func() bool {
		return !settings.Offline && gameNet.State() == gameNet.ConnectionConnected
	}

	connectionWarning := func() string {
		if settings.Offline {
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
		if joinCodeButton != nil {
			joinCodeButton.Text = "Join with Game Code"
		}
	}

	sendGameRequest := func(packet packets.C2SPacket) {
		if err := gameNet.Send(packet); err != nil {
			fmt.Printf("failed to send game request: %v\n", err)
		}
	}

	joinGame := func(code string) {
		sendGameRequest(&packets.C2SJoinGamePacket{GameCode: code})
	}

	panelProgress := func() float32 {
		return playSmoothstep(joinPanelProgress)
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
		if !multiplayerEnabled() && joinPanelOpen {
			joinPanelOpen = false
			focusCodeInput = false
			codeInput.Blur()
		}

		target := float32(0)
		if joinPanelOpen {
			target = 1
		}

		joinPanelProgress = playMoveTowards(
			joinPanelProgress,
			target,
			float32(time.Duration(deltaNano))/float32(joinPanelTransition),
		)

		if joinCodeButton != nil {
			if joinPanelOpen {
				joinCodeButton.Text = "Cancel Game Code"
			} else {
				joinCodeButton.Text = "Join with Game Code"
			}
		}

		if focusCodeInput && joinPanelProgress >= 0.98 && codeInput != nil {
			codeInput.Focus()
			focusCodeInput = false
		}
	})

	codeIsValid := func() bool {
		return codeInput != nil && len([]rune(codeInput.Text)) == gameCodeLength
	}

	codeInput = ui.Input().
		WithPlaceholderText("GAME CODE").
		WithMaxTextLength(gameCodeLength).
		WithCharset(gameCodeCharset).
		WithInputTransformer(strings.ToUpper).
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
			joinGame(codeInput.Text)
		})

	codeHelp := ui.Text().
		WithTextDynamic(func() string {
			if codeIsValid() {
				return "Press Enter or Join"
			}
			return "Enter a 6-character game code"
		}).
		WithTextSize(24).
		WithTextColor(rl.Gray).
		WithAnchors(anchor.Center, anchor.Center).
		WithRelativePosDynamic(func(el *ui.TextElement) vec.Vec2i {
			progress := panelProgress()
			return vec.Vec2i{X: 0, Y: 244 + int32((1-progress)*16)}
		}).
		WithVisibleDynamic(func(el *ui.TextElement) bool {
			return joinPanelProgress > 0
		}).
		WithOpacityDynamic(func(el *ui.TextElement) float32 {
			return panelProgress()
		})

	joinCodeButton = playMenuButton("Join with Game Code").
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

	return ui.Screen().
		AddChild(controller).
		AddChild(
			ui.Text().
				WithText("Play").
				WithTextSize(88).
				WithTextColor(rl.Black).
				WithAnchors(anchor.Center, anchor.Top).
				WithRelativePosDynamic(func(el *ui.TextElement) vec.Vec2i {
					return vec.Vec2i{X: 0, Y: max(int32(42), el.Parent.Size().Y/8)}
				}),
		).
		AddChild(
			ui.Text().
				WithText("Choose how to play").
				WithTextSize(28).
				WithTextColor(rl.Gray).
				WithAnchors(anchor.Center, anchor.Top).
				WithRelativePosDynamic(func(el *ui.TextElement) vec.Vec2i {
					return vec.Vec2i{X: 0, Y: max(int32(136), el.Parent.Size().Y/8+92)}
				}),
		).
		AddChild(
			playMenuButton("Play Solo").
				WithRelativePosDynamic(menuButtonPos(-100)).
				WithClick(OpenSoloGameCreation),
		).
		AddChild(
			playMenuButton("Host a Game").
				WithRelativePosDynamic(menuButtonPos(-28)).
				WithEnabledDynamic(func(el *ui.ButtonElement) bool {
					return multiplayerEnabled()
				}).
				WithClick(OpenHostGameCreation),
		).
		AddChild(
			playMenuButton("Join Random Game").
				WithRelativePosDynamic(menuButtonPos(44)).
				WithEnabledDynamic(func(el *ui.ButtonElement) bool {
					return multiplayerEnabled()
				}).
				WithClick(func() {
					joinGame("")
				}),
		).
		AddChild(joinCodeButton).
		AddChild(codeInput).
		AddChild(joinButton).
		AddChild(codeHelp).
		AddChild(
			ui.Text().
				WithTextDynamic(connectionWarning).
				WithTextSize(22).
				WithTextColor(rl.DarkGray).
				WithAnchors(anchor.Bottom, anchor.Bottom).
				WithRelativePos(vec.Vec2i{X: 0, Y: -96}).
				WithVisibleDynamic(func(el *ui.TextElement) bool {
					return !multiplayerEnabled() && joinPanelProgress <= 0
				}),
		).
		AddChild(
			playMenuButton("Try Reconnect").
				WithTextSize(26).
				WithSize(vec.Vec2i{X: 240, Y: 48}).
				WithAnchors(anchor.Bottom, anchor.Bottom).
				WithRelativePos(vec.Vec2i{X: 0, Y: -24}).
				WithVisibleDynamic(func(el *ui.ButtonElement) bool {
					return !settings.Offline && gameNet.State() == gameNet.ConnectionDisconnected && joinPanelProgress <= 0
				}).
				WithClick(gameNet.RetryConnection),
		).
		AddChild(backButton(func() {
			SetActiveScreen(MainScreenID)
		})).
		AddChild(
			ui.Vignette(),
		).
		WithExit(resetJoinPanel)
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

func backButton(click func()) *ui.ButtonElement {
	return ui.Button().
		WithText("Back").
		WithTextSize(40).
		WithPadding(8).
		WithOutlineWidth(4).
		WithForegroundColors(ui.ColorSet{
			Default: &rl.DarkGray,
		}).
		WithBackgroundColors(ui.ColorSet{
			Default: &rl.LightGray,
			Hover:   util.ColorAdd(rl.LightGray, 25),
			Click:   util.ColorAdd(rl.LightGray, 40),
		}).
		WithOutlineColors(ui.ColorSet{
			Default: &rl.Gray,
		}).
		WithAnchors(anchor.BottomLeft, anchor.BottomLeft).
		WithRelativePos(vec.Vec2i{X: 20, Y: -20}).
		WithClick(click)
}

func playMoveTowards(current, target, amount float32) float32 {
	if current < target {
		return min(current+amount, target)
	}
	return max(current-amount, target)
}

func playSmoothstep(value float32) float32 {
	value = max(float32(0), min(value, float32(1)))
	return value * value * (3 - 2*value)
}
