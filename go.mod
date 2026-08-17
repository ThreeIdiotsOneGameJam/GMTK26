module github.com/threeidiotsonegamejam/gmtk26

go 1.26.5

require (
	github.com/aquilax/go-perlin v1.1.0
	github.com/coder/websocket v1.8.15
	github.com/gen2brain/raylib-go/raylib v0.60.0
	github.com/go-text/typesetting v0.3.4
	github.com/google/uuid v1.6.0
)

require (
	github.com/ebitengine/purego v0.10.0 // indirect
	github.com/jupiterrider/ffi v0.7.0 // indirect
	golang.org/x/exp v0.0.0-20240506185415-9bf2ced13842 // indirect
	golang.org/x/mod v0.38.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/tools v0.48.0 // indirect
)

replace (
	github.com/ThreeIdiotsOneGameJam/Raylib-Go-Wasm/wasm-runtime => ./Raylib-Go-Wasm/wasm-runtime
	github.com/gen2brain/raylib-go/raygui => ./Raylib-Go-Wasm/raygui
)

// The raylib fork below is web-only (it pulls in wasm-ffi-go which does not
// build on desktop). The web targets in the Makefile enable it automatically,
// so you never need to edit this file by hand.
//replace github.com/gen2brain/raylib-go/raylib => ./Raylib-Go-Wasm/raylib

tool golang.org/x/tools/cmd/stringer
