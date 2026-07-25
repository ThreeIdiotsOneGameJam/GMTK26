package settings

import (
	"fmt"

	"github.com/threeidiotsonegamejam/gmtk26/src/storage"
)

const storageKey = "settings"

type SettingsStore struct {
	Offline bool `json:"offline"`
}

var Current = SettingsStore{}

func Load() error {
	var loaded SettingsStore
	found, err := storage.Load(storageKey, &loaded)
	if err != nil {
		return fmt.Errorf("load settings: %w", err)
	}
	if found {
		Current = loaded
	}

	return nil
}

func Save() error {
	if err := storage.Save(storageKey, Current); err != nil {
		return fmt.Errorf("save settings: %w", err)
	}

	return nil
}
