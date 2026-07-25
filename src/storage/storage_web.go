//go:build web

package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"syscall/js"

	"github.com/threeidiotsonegamejam/gmtk26/src/util/jsonutil"
)

var browserStorage = js.Global().Get("localStorage")

func Save[T any](key string, value T) (err error) {
	if browserStorage.IsUndefined() || browserStorage.IsNull() {
		return errors.New("localStorage is unavailable")
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode save data: %w", err)
	}

	// Browser storage operations can throw JavaScript exceptions.
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("localStorage setItem failed: %v", recovered)
		}
	}()

	browserStorage.Call("setItem", key, string(data))
	return nil
}

func Load[T any](key string, destination *T) (found bool, err error) {
	if browserStorage.IsUndefined() || browserStorage.IsNull() {
		return false, errors.New("localStorage is unavailable")
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			found = false
			err = fmt.Errorf("localStorage getItem failed: %v", recovered)
		}
	}()

	value := browserStorage.Call("getItem", key)
	if value.IsNull() || value.IsUndefined() {
		return false, nil
	}

	if err := jsonutil.DecodeStrict([]byte(value.String()), destination); err != nil {
		return false, fmt.Errorf("%w: %v", ErrInvalidData, err)
	}

	return true, nil
}

func Delete(key string) {
	if !browserStorage.IsUndefined() && !browserStorage.IsNull() {
		browserStorage.Call("removeItem", key)
	}
}
