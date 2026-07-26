package screens

import (
	"strings"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/global"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
	"github.com/threeidiotsonegamejam/gmtk26/src/util"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
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
	// Transitions intentionally pause screen updates, so clear transient hover
	// state here as well as in Screen.Update.
	global.TooltipText = ""
	ui.BeginInputFrame()

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
	transitionProgress = ui.MoveTowards(
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

	alpha := ui.Smoothstep(1 - transitionProgress)
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
	position := tooltipPosition(
		global.MousePosition.RoundToInt(),
		vec.Vec2i{X: int32(rl.GetRenderWidth()), Y: int32(rl.GetRenderHeight())},
		vec.Vec2i{X: bgW, Y: bgH},
	)

	rl.DrawRectangle(position.X, position.Y, bgW, bgH, util.ColorOpacity(rl.Black, 0.6))
	for i, line := range lines {
		rl.DrawText(line, position.X+pad, position.Y+pad+int32(i)*lineH, textSize, rl.White)
	}
}

func tooltipPosition(mouse, viewport, tooltip vec.Vec2i) vec.Vec2i {
	x := mouse.X - tooltip.X/2
	y := mouse.Y - tooltip.Y - 10
	if y < 0 {
		y = mouse.Y + 10
	}

	maxX := max(int32(0), viewport.X-tooltip.X)
	maxY := max(int32(0), viewport.Y-tooltip.Y)
	return vec.Vec2i{
		X: max(int32(0), min(x, maxX)),
		Y: max(int32(0), min(y, maxY)),
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
			if escShowingCountdownSettings {
				escShowingCountdownSettings = false
				return
			}
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
