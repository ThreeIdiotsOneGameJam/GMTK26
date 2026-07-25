package screens

import (
	"image/color"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/global"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
)

const (
	screenFadeOutDuration = 110 * time.Millisecond
	screenFadeInDuration  = 160 * time.Millisecond
)

var activeScreen *ui.ScreenElement
var pendingScreen *ui.ScreenElement
var transitionAlpha float32
var transitionColor color.RGBA

// Retained instances needed for overlays and return navigation.
var playScreen *ui.ScreenElement
var gameScreen *ui.ScreenElement
var escScreen *ui.ScreenElement
var matchmakingScreen *ui.ScreenElement

func init() {
	activeScreen = NewMainScreen()
	activeScreen.Enter()
	transitionColor = activeScreen.BackgroundColor
}

func GetActiveScreen() *ui.ScreenElement {
	return activeScreen
}

func SetActiveScreen(screen *ui.ScreenElement) {
	if screen == nil {
		panic("active screen must not be nil")
	}

	if screen == activeScreen && pendingScreen == nil {
		return
	}

	transitionColor = blendScreenColors(activeScreen.BackgroundColor, screen.BackgroundColor)
	pendingScreen = screen
}

// GoToPreviousScreen activates previousScreen, or Main when previous is nil.
func GoToPreviousScreen(previousScreen *ui.ScreenElement) {
	if previousScreen == nil {
		SetActiveScreen(NewMainScreen())
		return
	}
	SetActiveScreen(previousScreen)
}

func playScreenOrNew() *ui.ScreenElement {
	if playScreen != nil {
		return playScreen
	}
	return NewPlayScreen(NewMainScreen())
}

func IsTransitioning() bool {
	return pendingScreen != nil || transitionAlpha > 0
}

func Update(deltaNano int64) {
	if activeScreen == nil {
		return
	}

	if !IsTransitioning() {
		if IsEscScreenOpen() {
			// Update the game under the modal with input blocked so hover/click
			// state clears instead of freezing, then give the overlay a pass.
			global.UIModalBlocksInput = true
			global.UIBlocksWorldInput = true
			gameScreen.UpdateExcept(deltaNano, escScreen)
			global.UIModalBlocksInput = false
			global.UIBlocksWorldInput = false
			escScreen.Update(deltaNano)
			global.UIBlocksWorldInput = true
		} else {
			activeScreen.Update(deltaNano)
		}
	}

	delta := time.Duration(deltaNano)

	if pendingScreen != nil {
		transitionAlpha = moveTowards(
			transitionAlpha,
			1,
			float32(delta)/float32(screenFadeOutDuration),
		)

		if transitionAlpha >= 1 {
			activeScreen.Exit()
			activeScreen = pendingScreen
			pendingScreen = nil
			activeScreen.Enter()
		}
		return
	}

	transitionAlpha = moveTowards(
		transitionAlpha,
		0,
		float32(delta)/float32(screenFadeInDuration),
	)
}

func Draw() {
	if activeScreen == nil {
		return
	}

	activeScreen.Draw()

	if transitionAlpha <= 0 {
		return
	}

	alpha := smoothstep(transitionAlpha)
	overlay := transitionColor
	overlay.A = uint8(alpha * 255)
	rl.DrawRectangle(
		0,
		0,
		int32(rl.GetRenderWidth()),
		int32(rl.GetRenderHeight()),
		overlay,
	)
}

func blendScreenColors(from, to color.RGBA) color.RGBA {
	return color.RGBA{
		R: uint8((uint16(from.R) + uint16(to.R)) / 2),
		G: uint8((uint16(from.G) + uint16(to.G)) / 2),
		B: uint8((uint16(from.B) + uint16(to.B)) / 2),
		A: 255,
	}
}

func moveTowards(current, target, amount float32) float32 {
	if current < target {
		return min(current+amount, target)
	}
	return max(current-amount, target)
}

func smoothstep(value float32) float32 {
	value = max(float32(0), min(value, float32(1)))
	return value * value * (3 - 2*value)
}

func ToggleEscScreen() {
	if activeScreen != gameScreen || escScreen == nil {
		return
	}
	if escScreen.Visible() {
		HideEscScreen()
		return
	}
	escScreen.WithVisible(true)
}

func IsEscScreenOpen() bool {
	return activeScreen == gameScreen && escScreen != nil && escScreen.Visible()
}
