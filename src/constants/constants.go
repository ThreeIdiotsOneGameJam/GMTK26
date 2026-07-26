package constants

import "fmt"

const (
	GameName = "miniciv"
	AppName  = "miniciv"

	ServerPort            uint16 = 58008
	ServerBindHost               = "0.0.0.0"
	DefaultServerHost            = "gmtk26.ndmh.xyz"
	ClientWebSocketScheme        = "wss"
)

var FallbackServerHosts = []string{"localhost"}

const (
	WindowWidth  int32 = 1200
	WindowHeight int32 = 675

	ViewportWidth  = 640
	ViewportHeight = 360
)

func DefaultServerAddresses() []string {
	addrs := []string{fmt.Sprintf("%s:%d", DefaultServerHost, ServerPort)}
	for _, host := range FallbackServerHosts {
		addrs = append(addrs, fmt.Sprintf("%s:%d", host, ServerPort))
	}
	return addrs
}
