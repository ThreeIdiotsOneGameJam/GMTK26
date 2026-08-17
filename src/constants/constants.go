package constants

import "fmt"

const (
	GameName = "miniciv"
	AppName  = "miniciv"

	ServerPort     uint16 = 58008
	ServerBindHost        = "0.0.0.0"
)

type ServerAddr struct {
	Host   string
	Secure bool
}

func (a ServerAddr) URL() string {
	scheme := "ws"
	if a.Secure {
		scheme = "wss"
	}
	return fmt.Sprintf("%s://%s:%d", scheme, a.Host, ServerPort)
}

func DefaultServerAddrs() []ServerAddr {
	return []ServerAddr{
		{Host: "gmtk26.ndmh.xyz", Secure: true},
		{Host: "132.145.12.10", Secure: true},
		{Host: "141.11.62.108", Secure: true},
		{Host: "localhost", Secure: false},
	}
}

const (
	WindowWidth  int32 = 1200
	WindowHeight int32 = 675

	ViewportWidth  = 640
	ViewportHeight = 360
)
