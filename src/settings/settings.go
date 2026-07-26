package settings

import (
	"fmt"

	"github.com/threeidiotsonegamejam/gmtk26/src/storage"
)

const storageKey = "settings"

type SettingsStore struct {
	Offline        bool    `json:"offline"`
	SFXVolume      float32 `json:"sfx_volume"`
	MusicVolume    float32 `json:"music_volume"`
	AmbienceVolume float32 `json:"ambience_volume"`
	ReducedMotion  bool    `json:"reduced_motion"`
}

var Current = SettingsStore{
	SFXVolume:      0.5,
	MusicVolume:    0.5,
	AmbienceVolume: 0.5,
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

	clamped := Current
	clamped.SFXVolume = clamp01(Current.SFXVolume)
	clamped.MusicVolume = clamp01(Current.MusicVolume)
	clamped.AmbienceVolume = clamp01(Current.AmbienceVolume)

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
