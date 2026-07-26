//go:build darwin

package global

import (
	"os"
	"path/filepath"
)

var AssetDir = (func() string {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}

	exeDir := filepath.Dir(exe)
	if filepath.Base(exeDir) != "MacOS" {
		return "assets"
	}
	resources := filepath.Join(exeDir, "..", "Resources")

	return resources + "/assets"
})()
