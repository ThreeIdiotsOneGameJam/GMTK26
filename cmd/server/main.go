package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/threeidiotsonegamejam/gmtk26/src/constants"
	"github.com/threeidiotsonegamejam/gmtk26/src/net"
)

const (
	serverHostEnv = "SERVER_HOST"
	serverPortEnv = "SERVER_PORT"
)

type serverConfig struct {
	host string
	port uint16
}

func main() {
	config, err := parseServerConfig(os.Args[1:], os.LookupEnv, os.Stderr)
	if errors.Is(err, flag.ErrHelp) {
		return
	}
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("server starting")
	net.StartWebSocketServer(config.host, config.port)
}

func parseServerConfig(
	args []string,
	lookupEnv func(string) (string, bool),
	output io.Writer,
) (serverConfig, error) {
	config := serverConfig{
		host: constants.ServerBindHost,
		port: constants.ServerPort,
	}

	if host, ok := lookupEnv(serverHostEnv); ok {
		config.host = host
	}

	port := uint64(config.port)
	var environmentPortError error
	if value, ok := lookupEnv(serverPortEnv); ok {
		parsedPort, err := parsePort(value)
		if err != nil {
			environmentPortError = fmt.Errorf("invalid %s: %w", serverPortEnv, err)
		} else {
			port = parsedPort
		}
	}

	flags := flag.NewFlagSet("game-server", flag.ContinueOnError)
	flags.SetOutput(output)
	flags.StringVar(
		&config.host,
		"host",
		config.host,
		"host or IP address to listen on (env: "+serverHostEnv+")",
	)
	flags.Uint64Var(
		&port,
		"port",
		port,
		"TCP port to listen on (env: "+serverPortEnv+")",
	)

	if err := flags.Parse(args); err != nil {
		return serverConfig{}, err
	}
	portFlagSet := false
	flags.Visit(func(currentFlag *flag.Flag) {
		if currentFlag.Name == "port" {
			portFlagSet = true
		}
	})
	if environmentPortError != nil && !portFlagSet {
		return serverConfig{}, environmentPortError
	}
	if flags.NArg() != 0 {
		return serverConfig{}, fmt.Errorf("unexpected arguments: %s", strings.Join(flags.Args(), " "))
	}
	if strings.TrimSpace(config.host) == "" {
		return serverConfig{}, errors.New("host must not be empty")
	}
	if port == 0 || port > 65535 {
		return serverConfig{}, fmt.Errorf("port must be between 1 and 65535, got %d", port)
	}

	config.port = uint16(port)
	return config, nil
}

func parsePort(value string) (uint64, error) {
	port, err := strconv.ParseUint(value, 10, 16)
	if err != nil || port == 0 {
		return 0, fmt.Errorf("port must be an integer between 1 and 65535, got %q", value)
	}
	return port, nil
}
