package constants

import "fmt"

const (
	GameName = "miniciv"
	AppName  = "miniciv"

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
