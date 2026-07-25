package ui

import "image/color"

// UI colors from Indigo Glass
// (https://github.com/JohnRebellion/indigo-glass — indigo heritage variant).
var (
	PaletteIndigo      = color.RGBA{R: 0x5E, G: 0x6A, B: 0xD2, A: 255} // #5E6AD2 primary
	PaletteIndigoHover = color.RGBA{R: 0x81, G: 0x8C, B: 0xF8, A: 255} // #818CF8 hover/focus
	PaletteViolet      = color.RGBA{R: 0xA7, G: 0x8B, B: 0xFA, A: 255} // #A78BFA accent
	PaletteBase        = color.RGBA{R: 0x0F, G: 0x0F, B: 0x12, A: 255} // #0F0F12
	PaletteSurface     = color.RGBA{R: 0x1C, G: 0x1C, B: 0x21, A: 255} // #1C1C21
	PaletteSurfaceUp   = color.RGBA{R: 0x1F, G: 0x20, B: 0x28, A: 255} // #1F2028
	PaletteBorder      = color.RGBA{R: 0x2A, G: 0x2A, B: 0x32, A: 255} // surface-family border
	PaletteText        = color.RGBA{R: 0xF8, G: 0xF8, B: 0xF8, A: 255} // #F8F8F8
	PaletteTextMuted   = color.RGBA{R: 0x6B, G: 0x72, B: 0x80, A: 255} // #6B7280
	// Slate-300 — secondary labels over the decorative menu backdrop
	// (Indigo Glass muted is for near-black surfaces and reads too dim there).
	PaletteTextSecondary = color.RGBA{R: 0xCB, G: 0xD5, B: 0xE1, A: 255} // #CBD5E1
	PaletteIndigoDim     = color.RGBA{R: 0x3F, G: 0x45, B: 0x80, A: 255} // disabled indigo
	PaletteIndigoPress   = color.RGBA{R: 0x4F, G: 0x5A, B: 0xB8, A: 255} // pressed indigo
	PaletteNegative      = color.RGBA{R: 0xED, G: 0x25, B: 0x4E, A: 255} // #ED254E
)

func ptrColor(c color.RGBA) *color.RGBA {
	return &c
}
