//go:build !web

package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/threeidiotsonegamejam/gmtk26/src/util/jsonutil"
)

const appName = "3I1GJ-GMTK26"

func filePath(key string) (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("find user config directory: %w", err)
	}

	dir := filepath.Join(configDir, appName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create save directory: %w", err)
	}

	return filepath.Join(dir, key+".json"), nil
}

func Save[T any](key string, value T) error {
	path, err := filePath(key)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode save data: %w", err)
	}

	// Write to a temporary file first to reduce save corruption risk.
	tempPath := path + ".tmp"

	if err := os.WriteFile(tempPath, data, 0o600); err != nil {
		return fmt.Errorf("write temporary save file: %w", err)
	}

	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("replace save file: %w", err)
	}

	return nil
}

func Load[T any](key string, destination *T) (bool, error) {
	path, err := filePath(key)
	if err != nil {
		return false, err
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("read save file: %w", err)
	}

	if err := jsonutil.DecodeStrict(data, destination); err != nil {
		return false, fmt.Errorf("%w: %v", ErrInvalidData, err)
	}

	return true, nil
}

func Delete(key string) error {
	path, err := filePath(key)
	if err != nil {
		return err
	}

	err = os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	return err
}
