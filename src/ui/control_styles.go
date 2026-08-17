package ui

var (
	DefaultTrackColors = ColorSet{
		Default:  ptrColor(PaletteSurface),
		Hover:    ptrColor(PaletteSurfaceUp),
		Click:    ptrColor(PaletteSurfaceUp),
		Disabled: ptrColor(PaletteBase),
	}
	DefaultActiveTrackColors = ColorSet{
		Default:  ptrColor(PaletteIndigo),
		Hover:    ptrColor(PaletteIndigoHover),
		Click:    ptrColor(PaletteViolet),
		Disabled: ptrColor(PaletteIndigoDim),
	}
	DefaultThumbColors = ColorSet{
		Default:  ptrColor(PaletteText),
		Hover:    ptrColor(PaletteText),
		Click:    ptrColor(PaletteText),
		Disabled: ptrColor(PaletteTextMuted),
	}
	DefaultTrackOutlineColors = ColorSet{
		Default:  ptrColor(PaletteBorder),
		Disabled: ptrColor(PaletteBase),
	}
)
