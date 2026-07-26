package screens

import (
	"strings"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/global"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
	"github.com/threeidiotsonegamejam/gmtk26/src/util"
)

const (
	screenCrossfadeDuration = 250 * time.Millisecond
	// Capturing the outgoing framebuffer can stall WebGL for a long frame.
	// Do not let that stall skip most (or all) of the visual transition.
	screenCrossfadeMaxStep = time.Second / 30
)

var activeScreen *ui.ScreenElement
var pendingScreen *ui.ScreenElement
var transitionProgress float32
var transitionCanceling bool
var transitionReady bool
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
	transitionReady = false
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
	transitionReady = false
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
	// The first frame must show the captured source at full opacity. In
	// particular, do not advance using the frame time spent in glReadPixels.
	if !transitionReady {
		return
	}

	target := float32(1)
	if transitionCanceling {
		target = 0
	}
	step := min(time.Duration(deltaNano), screenCrossfadeMaxStep)
	transitionProgress = moveTowards(
		transitionProgress,
		target,
		float32(step)/float32(screenCrossfadeDuration),
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
	transitionReady = false
	releaseTransitionSource()
}

func Draw() {
	if activeScreen == nil {
		return
	}

	if pendingScreen == nil {
		activeScreen.Draw()
		drawTooltip()
		return
	}

	activeScreen.Draw()

	w := int32(rl.GetRenderWidth())
	h := int32(rl.GetRenderHeight())
	if w <= 0 || h <= 0 {
		drawTooltip()
		return
	}
	captured := ensureTransitionSource(w, h)
	transitionReady = true
	if !captured {
		// The source is already visible. If capture fails, show the destination
		// only while moving forward and keep the source visible while canceling.
		if !transitionCanceling {
			pendingScreen.Draw()
		}
		drawTooltip()
		return
	}

	// Draw the destination directly to the window. Screens such as the game
	// use their own render textures, which cannot be nested in raylib.
	pendingScreen.Draw()

	alpha := smoothstep(1 - transitionProgress)
	opacity := uint8(alpha * 255)
	// Draw an explicitly premultiplied source. The default alpha blend also
	// reduces framebuffer alpha during a fade, allowing browsers to composite
	// the dark vignette against the page's white background.
	tint := rl.Color{R: opacity, G: opacity, B: opacity, A: opacity}
	src := rl.Rectangle{X: 0, Y: 0, Width: float32(w), Height: float32(h)}
	dst := rl.Rectangle{X: 0, Y: 0, Width: float32(w), Height: float32(h)}
	rl.BeginBlendMode(rl.BlendAlphaPremultiply)
	rl.DrawTexturePro(transitionSource, src, dst, rl.Vector2{}, 0, tint)
	rl.EndBlendMode()

	drawTooltip()
}

func drawTooltip() {
	text := global.TooltipText
	if text == "" {
		return
	}

	lines := strings.Split(text, "\n")
	textSize := int32(18)
	lineH := textSize + 2
	pad := int32(6)

	textW := int32(0)
	for _, line := range lines {
		w := rl.MeasureText(line, textSize)
		if w > textW {
			textW = w
		}
	}
	textH := int32(len(lines)) * lineH
	bgW := textW + pad*2
	bgH := textH + pad*2
	mx := int32(global.MousePosition.X)
	my := int32(global.MousePosition.Y)
	x := mx - bgW/2
	y := my - bgH - 10

	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = my + 10
	}

	rl.DrawRectangle(x, y, bgW, bgH, util.ColorOpacity(rl.Black, 0.6))
	for i, line := range lines {
		rl.DrawText(line, x+pad, y+pad+int32(i)*lineH, textSize, rl.White)
	}
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
	transitionReady = false
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
