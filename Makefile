WEB_DIR := ./Raylib-Go-Wasm/index
SERVER_BIN := ./server$(shell go env GOEXE)
GOMOD := go.mod

# The raylib fork is web-only; this enables/disables its replace directive.
web_replace_on:
	sed -i 's|^//replace github.com/gen2brain/raylib-go/raylib|replace github.com/gen2brain/raylib-go/raylib|' $(GOMOD)

web_replace_off:
	sed -i 's|^replace github.com/gen2brain/raylib-go/raylib|//replace github.com/gen2brain/raylib-go/raylib|' $(GOMOD)

.PHONY: run_desktop build_desktop run_web build_web server clean web_replace_on web_replace_off

## Desktop (native) ##
build_desktop: web_replace_off
	go mod tidy
	go build -o ./bin/desktop .

run_desktop: build_desktop
	./bin/desktop

## Web (WASM) ##
build_web: web_replace_on
	go mod tidy
	GOOS=js GOARCH=wasm go build -tags web -o $(WEB_DIR)/main.wasm .
	cp "$(shell go env GOROOT)/lib/wasm/wasm_exec.js" $(WEB_DIR)/wasm_exec.js
	$(MAKE) web_replace_off

# Build the static server once (only needed the first time)
server:
	go build -o $(SERVER_BIN) ./Raylib-Go-Wasm/server/server.go

# Build the wasm and run the local server on http://localhost:8080
run_web: build_web server
	$(SERVER_BIN)

clean:
	rm -rf ./bin $(SERVER_BIN) $(WEB_DIR)/main.wasm $(WEB_DIR)/wasm_exec.js
