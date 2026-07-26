package settings

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/threeidiotsonegamejam/gmtk26/src/storage"
)

const storageKey = "settings"

type CountdownAnchor string

const (
	CountdownGridSize      = int32(5)
	CountdownCenter        = CountdownAnchor("2_2")
	DefaultCountdownAnchor = CountdownAnchor("2_3")
)

const (
	DefaultCountdownScale = float32(0.75)
	MinCountdownScale     = float32(0.25)
	MaxCountdownScale     = float32(1.5)
)

type SettingsStore struct {
	Offline         bool            `json:"offline"`
	SFXVolume       float32         `json:"sfx_volume"`
	MusicVolume     float32         `json:"music_volume"`
	AmbienceVolume  float32         `json:"ambience_volume"`
	ReducedMotion   bool            `json:"reduced_motion"`
	CountdownScale  float32         `json:"countdown_scale"`
	CountdownAnchor CountdownAnchor `json:"countdown_anchor"`
}

var Current = SettingsStore{
	SFXVolume:       0.5,
	MusicVolume:     0.5,
	AmbienceVolume:  0.5,
	CountdownScale:  DefaultCountdownScale,
	CountdownAnchor: DefaultCountdownAnchor,
}

func Load() error {
	var loaded SettingsStore
	found, err := storage.Load(storageKey, &loaded)
	if err != nil {
		return fmt.Errorf("load settings: %w", err)
	}
	if found {
		Current = loaded
	}

	clamped := normalized(Current)

	if clamped != Current {
		Current = clamped
		if err := Save(); err != nil {
			return fmt.Errorf("rewrite clamped settings: %w", err)
		}
	}

	return nil
}

func Save() error {
	if err := storage.Save(storageKey, Current); err != nil {
		return fmt.Errorf("save settings: %w", err)
	}

	return nil
}

func clamp01(v float32) float32 {
	return min(max(v, 0), 1)
}

func normalized(value SettingsStore) SettingsStore {
	value.SFXVolume = clamp01(value.SFXVolume)
	value.MusicVolume = clamp01(value.MusicVolume)
	value.AmbienceVolume = clamp01(value.AmbienceVolume)

	// Zero and an empty anchor identify settings saved by versions before
	// countdown customization existed.
	if value.CountdownScale == 0 {
		value.CountdownScale = DefaultCountdownScale
	} else {
		value.CountdownScale = min(
			max(value.CountdownScale, MinCountdownScale),
			MaxCountdownScale,
		)
	}
	column, row, valid := CountdownAnchorGridPosition(value.CountdownAnchor)
	if !valid {
		value.CountdownAnchor = DefaultCountdownAnchor
	} else {
		// Canonicalize legacy 3x3 values while preserving their nearest
		// equivalent position in the new 5x5 grid.
		value.CountdownAnchor = CountdownAnchorAt(column, row)
	}

	return value
}

func CountdownAnchorAt(column, row int32) CountdownAnchor {
	if column < 0 || column >= CountdownGridSize ||
		row < 0 || row >= CountdownGridSize {
		return CountdownCenter
	}
	return CountdownAnchor(
		strconv.FormatInt(int64(column), 10) + "_" +
			strconv.FormatInt(int64(row), 10),
	)
}

func CountdownAnchorGridPosition(value CountdownAnchor) (column, row int32, valid bool) {
	// Settings written by the previous 3x3 selector map to the nearest 5x5
	// position: 25/50/75% becomes 30/50/70%.
	switch value {
	case "top_left":
		return 1, 1, true
	case "top":
		return 2, 1, true
	case "top_right":
		return 3, 1, true
	case "left":
		return 1, 2, true
	case "center":
		return 2, 2, true
	case "right":
		return 3, 2, true
	case "bottom_left":
		return 1, 3, true
	case "bottom":
		return 2, 3, true
	case "bottom_right":
		return 3, 3, true
	}

	parts := strings.Split(string(value), "_")
	if len(parts) != 2 {
		return 0, 0, false
	}
	parsedColumn, columnErr := strconv.ParseInt(parts[0], 10, 32)
	parsedRow, rowErr := strconv.ParseInt(parts[1], 10, 32)
	column, row = int32(parsedColumn), int32(parsedRow)
	if columnErr != nil || rowErr != nil ||
		column < 0 || column >= CountdownGridSize ||
		row < 0 || row >= CountdownGridSize {
		return 0, 0, false
	}
	return column, row, true
}
