# GMTK26

gaylib

## Development

### Setup Web

1. Copy Golang wasm runtime (only needs to be copied once)
   `cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" ./Raylib-Go-Wasm/index/wasm_exec.js`
2. Compile the server (only needs to be compiled once) - add .exe on Windows
   `go build -o bin/server ./Raylib-Go-Wasm/server/server.go`
3. Run the server - add .exe on Windows
   `./bin/server`
4. Visit http://localhost:8080
5. In another terminal run [`air`](https://github.com/air-verse/air) to automatically rebuild your code.

### Docs

https://github.com/gen2brain/raylib-go
https://github.com/BrownNPC/Raylib-Go-Wasm
