package main

import (
	"io"
	"strings"
	"testing"

	"github.com/threeidiotsonegamejam/gmtk26/src/constants"
)

func TestParseServerConfigDefaults(t *testing.T) {
	config, err := parseServerConfig(nil, noEnvironment, io.Discard)
	if err != nil {
		t.Fatal(err)
	}

	if config.host != constants.ServerBindHost {
		t.Errorf("host = %q, want %q", config.host, constants.ServerBindHost)
	}
	if config.port != constants.ServerPort {
		t.Errorf("port = %d, want %d", config.port, constants.ServerPort)
	}
}

func TestParseServerConfigEnvironment(t *testing.T) {
	environment := map[string]string{
		serverHostEnv: "127.0.0.1",
		serverPortEnv: "9000",
	}

	config, err := parseServerConfig(nil, mapEnvironment(environment), io.Discard)
	if err != nil {
		t.Fatal(err)
	}

	if config.host != "127.0.0.1" {
		t.Errorf("host = %q, want %q", config.host, "127.0.0.1")
	}
	if config.port != 9000 {
		t.Errorf("port = %d, want %d", config.port, 9000)
	}
}

func TestParseServerConfigFlagsOverrideEnvironment(t *testing.T) {
	environment := map[string]string{
		serverHostEnv: "127.0.0.1",
		serverPortEnv: "not-a-port",
	}

	config, err := parseServerConfig(
		[]string{"--host", "::", "--port", "10000"},
		mapEnvironment(environment),
		io.Discard,
	)
	if err != nil {
		t.Fatal(err)
	}

	if config.host != "::" {
		t.Errorf("host = %q, want %q", config.host, "::")
	}
	if config.port != 10000 {
		t.Errorf("port = %d, want %d", config.port, 10000)
	}
}

func TestParseServerConfigRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		environment map[string]string
		wantError   string
	}{
		{
			name:        "invalid environment port",
			environment: map[string]string{serverPortEnv: "not-a-port"},
			wantError:   serverPortEnv,
		},
		{
			name:      "out of range flag port",
			args:      []string{"--port", "65536"},
			wantError: "between 1 and 65535",
		},
		{
			name:      "empty flag host",
			args:      []string{"--host", ""},
			wantError: "host must not be empty",
		},
		{
			name:      "positional arguments",
			args:      []string{"extra"},
			wantError: "unexpected arguments",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := parseServerConfig(
				test.args,
				mapEnvironment(test.environment),
				io.Discard,
			)
			if err == nil {
				t.Fatal("expected an error")
			}
			if !strings.Contains(err.Error(), test.wantError) {
				t.Errorf("error = %q, want it to contain %q", err, test.wantError)
			}
		})
	}
}

func noEnvironment(string) (string, bool) {
	return "", false
}

func mapEnvironment(values map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}
