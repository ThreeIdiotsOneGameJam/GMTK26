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

### Game server

Build and run the multiplayer server:

```sh
make run_server
```

It listens on `0.0.0.0:58008` by default. The address can be configured with
command-line flags or environment variables; command-line flags take
precedence:

```sh
SERVER_HOST=127.0.0.1 SERVER_PORT=9000 ./bin/game-server
./bin/game-server --host 127.0.0.1 --port 9000
```

Build and run the Docker image:

```sh
docker build -t gmtk26-server .
docker run --rm -p 58008:58008 gmtk26-server
```

Use `-e SERVER_PORT=9000 -p 9000:9000` to run the container on a different
port.

The **Build and publish server image** GitHub Actions workflow can be started
manually. It runs on a GitHub-hosted runner, caches Docker layers, and publishes
the image to `ghcr.io/threeidiotsonegamejam/gmtk26`.

For a Compose-based production setup, including GHCR authentication, firewall
configuration, TLS, upgrades, and rollback, see the
[server deployment guide](docs/server-deployment.md).
