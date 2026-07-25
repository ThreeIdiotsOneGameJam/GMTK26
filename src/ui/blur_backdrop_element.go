package ui

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/util/vec"
)

const blurKawaseIterations = 3

func BlurBackdrop() *BlurBackdropElement {
	el := &BlurBackdropElement{
		Tint:   color.RGBA{R: 40, G: 40, B: 40, A: 100},
		Offset: 1.5,
	}
	el.BaseElement = NewBaseElement(el)
	return el.WithSizeDynamic(func(_ *BlurBackdropElement) vec.Vec2i {
		return vec.Vec2i{X: int32(rl.GetRenderWidth()), Y: int32(rl.GetRenderHeight())}
	})
}

var blurShader rl.Shader

type BlurBackdropElement struct {
	BaseElement[*BlurBackdropElement]

	Tint   color.RGBA
	Offset float32

	capture  rl.Texture2D
	captureW int32
	captureH int32

	// Dual Kawase pyramid: downs[i] is half the resolution of downs[i-1].
	downs []rl.RenderTexture2D
}

func (el *BlurBackdropElement) WithTint(tint color.RGBA) *BlurBackdropElement {
	el.Tint = tint
	return el
}

// WithBlurRadius sets the Dual Kawase sample offset (typically 1.0–2.5).
func (el *BlurBackdropElement) WithBlurRadius(radius float32) *BlurBackdropElement {
	el.Offset = radius
	return el
}

func (el *BlurBackdropElement) WithOffset(offset float32) *BlurBackdropElement {
	el.Offset = offset
	return el
}

func (el *BlurBackdropElement) Release() {
	if rl.IsTextureValid(el.capture) {
		rl.UnloadTexture(el.capture)
		el.capture = rl.Texture2D{}
	}
	for i := range el.downs {
		if rl.IsRenderTextureValid(el.downs[i]) {
			rl.UnloadRenderTexture(el.downs[i])
		}
	}
	el.downs = nil
	el.captureW = 0
	el.captureH = 0
}

func (el *BlurBackdropElement) prepare() {
	if !rl.IsShaderValid(blurShader) {
		blurShader = rl.LoadShader("assets/shaders/base.vert", "assets/shaders/blur.frag")
	}
}

func (el *BlurBackdropElement) draw() {
	w := int32(rl.GetRenderWidth())
	h := int32(rl.GetRenderHeight())
	if w <= 0 || h <= 0 {
		return
	}

	// Game keeps simulating/drawing under the overlay, so recapture every frame.
	el.captureScreen(w, h)
	if !rl.IsTextureValid(el.capture) {
		el.drawTint(w, h)
		return
	}

	if !rl.IsShaderValid(blurShader) {
		rl.DrawTexture(el.capture, 0, 0, rl.White)
		el.drawTint(w, h)
		return
	}

	el.ensurePyramid(w, h)
	if len(el.downs) == 0 {
		rl.DrawTexture(el.capture, 0, 0, rl.White)
		el.drawTint(w, h)
		return
	}

	halfPixelLoc := rl.GetShaderLocation(blurShader, "halfPixel")
	offsetLoc := rl.GetShaderLocation(blurShader, "offset")
	upsampleLoc := rl.GetShaderLocation(blurShader, "upsample")
	rl.SetShaderValue(blurShader, offsetLoc, []float32{el.Offset}, rl.ShaderUniformFloat)

	// Downsample chain: full capture -> 1/2 -> 1/4 -> ...
	srcTex := el.capture
	srcW, srcH := w, h
	fromRT := false
	for i := range el.downs {
		dst := el.downs[i]
		dstW := dst.Texture.Width
		dstH := dst.Texture.Height
		rl.SetShaderValue(
			blurShader,
			halfPixelLoc,
			[]float32{0.5 / float32(dstW), 0.5 / float32(dstH)},
			rl.ShaderUniformVec2,
		)
		rl.SetShaderValue(blurShader, upsampleLoc, []float32{0}, rl.ShaderUniformFloat)
		el.blitKawase(dst, srcTex, srcW, srcH, fromRT)
		srcTex = dst.Texture
		srcW, srcH = dstW, dstH
		fromRT = true
	}

	// Upsample chain back toward full resolution, final pass to the screen.
	for i := len(el.downs) - 2; i >= 0; i-- {
		dst := el.downs[i]
		dstW := dst.Texture.Width
		dstH := dst.Texture.Height
		rl.SetShaderValue(
			blurShader,
			halfPixelLoc,
			[]float32{0.5 / float32(srcW), 0.5 / float32(srcH)},
			rl.ShaderUniformVec2,
		)
		rl.SetShaderValue(blurShader, upsampleLoc, []float32{1}, rl.ShaderUniformFloat)
		el.blitKawase(dst, srcTex, srcW, srcH, true)
		srcTex = dst.Texture
		srcW, srcH = dstW, dstH
	}

	rl.SetShaderValue(
		blurShader,
		halfPixelLoc,
		[]float32{0.5 / float32(srcW), 0.5 / float32(srcH)},
		rl.ShaderUniformVec2,
	)
	rl.SetShaderValue(blurShader, upsampleLoc, []float32{1}, rl.ShaderUniformFloat)
	rl.BeginShaderMode(blurShader)
	src := rl.Rectangle{X: 0, Y: 0, Width: float32(srcW), Height: float32(-srcH)}
	dst := rl.Rectangle{X: 0, Y: 0, Width: float32(w), Height: float32(h)}
	rl.DrawTexturePro(srcTex, src, dst, rl.Vector2{}, 0, rl.White)
	rl.EndShaderMode()

	el.drawTint(w, h)
}

func (el *BlurBackdropElement) blitKawase(
	dst rl.RenderTexture2D,
	src rl.Texture2D,
	srcW, srcH int32,
	fromRT bool,
) {
	rl.BeginTextureMode(dst)
	rl.ClearBackground(rl.Blank)
	rl.BeginShaderMode(blurShader)

	srcRect := rl.Rectangle{X: 0, Y: 0, Width: float32(srcW), Height: float32(srcH)}
	if fromRT {
		srcRect.Height = -float32(srcH)
	}
	dstRect := rl.Rectangle{
		X:      0,
		Y:      0,
		Width:  float32(dst.Texture.Width),
		Height: float32(dst.Texture.Height),
	}
	rl.DrawTexturePro(src, srcRect, dstRect, rl.Vector2{}, 0, rl.White)

	rl.EndShaderMode()
	rl.EndTextureMode()
}

func (el *BlurBackdropElement) drawTint(w, h int32) {
	if el.Tint.A == 0 {
		return
	}
	rl.DrawRectangle(0, 0, w, h, el.Tint)
}

func (el *BlurBackdropElement) ensurePyramid(w, h int32) {
	if len(el.downs) == blurKawaseIterations &&
		rl.IsRenderTextureValid(el.downs[0]) &&
		el.downs[0].Texture.Width == max(w/2, 1) &&
		el.downs[0].Texture.Height == max(h/2, 1) {
		return
	}

	for i := range el.downs {
		if rl.IsRenderTextureValid(el.downs[i]) {
			rl.UnloadRenderTexture(el.downs[i])
		}
	}
	el.downs = make([]rl.RenderTexture2D, 0, blurKawaseIterations)

	pw, ph := w, h
	for range blurKawaseIterations {
		pw = max(pw/2, 1)
		ph = max(ph/2, 1)
		rt := rl.LoadRenderTexture(pw, ph)
		if !rl.IsRenderTextureValid(rt) {
			el.Release()
			return
		}
		rl.SetTextureFilter(rt.Texture, rl.FilterBilinear)
		el.downs = append(el.downs, rt)
	}
}

func (el *BlurBackdropElement) captureScreen(w, h int32) {
	img := rl.LoadImageFromScreen()
	if img == nil {
		return
	}
	defer rl.UnloadImage(img)

	if rl.IsTextureValid(el.capture) && el.captureW == w && el.captureH == h {
		pixels := rl.LoadImageColors(img)
		rl.UpdateTexture(el.capture, pixels)
		rl.UnloadImageColors(pixels)
		return
	}

	if rl.IsTextureValid(el.capture) {
		rl.UnloadTexture(el.capture)
		el.capture = rl.Texture2D{}
	}
	el.capture = rl.LoadTextureFromImage(img)
	if !rl.IsTextureValid(el.capture) {
		return
	}
	rl.SetTextureFilter(el.capture, rl.FilterBilinear)
	el.captureW = w
	el.captureH = h
}
