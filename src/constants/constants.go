package constants

import "fmt"

const (
	GameName = "Game"
	AppName  = "3I1GJ-GMTK26"

	ServerPort        uint16 = 58008
	ServerBindHost           = "0.0.0.0"
	DefaultServerHost        = "localhost"

	WindowWidth  int32 = 1200
	WindowHeight int32 = 675

	ViewportWidth  = 640
	ViewportHeight = 360
)

func DefaultServerAddress() string {
	return fmt.Sprintf("%s:%d", DefaultServerHost, ServerPort)
}
