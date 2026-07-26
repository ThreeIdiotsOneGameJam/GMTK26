package screens

import (
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/global"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
)

const screenCrossfadeDuration = 250 * time.Millisecond

var activeScreen *ui.ScreenElement
var pendingScreen *ui.ScreenElement
var transitionProgress float32
var transitionCanceling bool
var transitionSource rl.Texture2D

// Retained instances needed for overlays and return navigation.
var playScreen *ui.ScreenElement
var gameScreen *ui.ScreenElement
var escScreen *ui.ScreenElement
var matchmakingScreen *ui.ScreenElement

func init() {
	activeScreen = NewMainScreen()
	activeScreen.Enter()
}

func GetActiveScreen() *ui.ScreenElement {
	return activeScreen
}

func screenIsActiveOrPending(screen *ui.ScreenElement) bool {
	return screen != nil && (activeScreen == screen || pendingScreen == screen)
}

func SetActiveScreen(screen *ui.ScreenElement) {
	if screen == nil {
		panic("active screen must not be nil")
	}

	if screen == activeScreen {
		cancelTransition()
		return
	}
	if screen == pendingScreen {
		transitionCanceling = false
		return
	}

	if pendingScreen != nil {
		pendingScreen.Exit()
	}
	releaseTransitionSource()

	pendingScreen = screen
	transitionProgress = 0
	transitionCanceling = false
	pendingScreen.Enter()
}

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
	return pendingScreen != nil
}

func cancelTransition() {
	if pendingScreen == nil {
		return
	}
	transitionCanceling = true
}

func finishCanceledTransition() {
	pendingScreen.Exit()
	pendingScreen = nil
	transitionProgress = 0
	transitionCanceling = false
	releaseTransitionSource()
}

func Update(deltaNano int64) {
	if activeScreen == nil {
		return
	}

	if !IsTransitioning() {
		if IsEscScreenOpen() {
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

	if pendingScreen == nil {
		return
	}

	target := float32(1)
	if transitionCanceling {
		target = 0
	}
	transitionProgress = moveTowards(
		transitionProgress,
		target,
		float32(time.Duration(deltaNano))/float32(screenCrossfadeDuration),
	)

	if transitionCanceling {
		if transitionProgress <= 0 {
			finishCanceledTransition()
		}
		return
	}

	if transitionProgress < 1 {
		return
	}

	activeScreen.Exit()
	activeScreen = pendingScreen
	pendingScreen = nil
	transitionProgress = 0
	transitionCanceling = false
	releaseTransitionSource()
}

func Draw() {
	if activeScreen == nil {
		return
	}

	if pendingScreen == nil {
		activeScreen.Draw()
		return
	}

	activeScreen.Draw()

	w := int32(rl.GetRenderWidth())
	h := int32(rl.GetRenderHeight())
	if w <= 0 || h <= 0 {
		return
	}
	if !ensureTransitionSource(w, h) {
		// The source is already visible. If capture fails, show the destination
		// only while moving forward and keep the source visible while canceling.
		if !transitionCanceling {
			pendingScreen.Draw()
		}
		return
	}

	// Draw the destination directly to the window. Screens such as the game
	// use their own render textures, which cannot be nested in raylib.
	pendingScreen.Draw()

	alpha := smoothstep(1 - transitionProgress)
	tint := rl.Color{R: 255, G: 255, B: 255, A: uint8(alpha * 255)}
	src := rl.Rectangle{X: 0, Y: 0, Width: float32(w), Height: float32(h)}
	dst := rl.Rectangle{X: 0, Y: 0, Width: float32(w), Height: float32(h)}
	rl.DrawTexturePro(transitionSource, src, dst, rl.Vector2{}, 0, tint)
}

func ensureTransitionSource(w, h int32) bool {
	if rl.IsTextureValid(transitionSource) &&
		transitionSource.Width == w &&
		transitionSource.Height == h {
		return true
	}

	releaseTransitionSource()
	flushTransitionSourceDraws()
	image := rl.LoadImageFromScreen()
	if image == nil {
		return false
	}
	defer rl.UnloadImage(image)

	transitionSource = rl.LoadTextureFromImage(image)
	return rl.IsTextureValid(transitionSource)
}

func releaseTransitionSource() {
	if !rl.IsTextureValid(transitionSource) {
		return
	}
	rl.UnloadTexture(transitionSource)
	transitionSource = rl.Texture2D{}
}

func Shutdown() {
	if pendingScreen != nil {
		pendingScreen.Exit()
		pendingScreen = nil
	}
	if activeScreen != nil {
		activeScreen.Exit()
		activeScreen = nil
	}
	transitionProgress = 0
	transitionCanceling = false
	releaseTransitionSource()
	HideEscScreen()
	gameWorld.Renderer.Unload()
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

func HandleEscape() {
	if pendingScreen != nil {
		if gameLeaveTransition {
			return
		}
		cancelTransition()
		return
	}
	if activeScreen == gameScreen {
		ToggleEscScreen()
		return
	}
	if activeScreen != nil && activeScreen.OnBack != nil {
		activeScreen.OnBack()
	}
}

func ToggleEscScreen() {
	if activeScreen != gameScreen || escScreen == nil {
		return
	}
	if escScreen.Visible() {
		if escShowingSettings {
			escShowingSettings = false
			return
		}
		HideEscScreen()
		return
	}
	escScreen.WithVisible(true)
}

func IsEscScreenOpen() bool {
	return activeScreen == gameScreen && escScreen != nil && escScreen.Visible()
}
