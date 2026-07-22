package screens

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
)

type ScreenID int

const (
	MainScreenID ScreenID = iota
	GameScreenID
	SettingsScreenID
)

var screenMap map[ScreenID]*ui.Screen
var activeScreen *ui.Screen

func init() {
	screenMap = map[ScreenID]*ui.Screen{
		MainScreenID:     &MainScreen,
		GameScreenID:     &GameScreen,
		SettingsScreenID: &SettingsScreen,
	}
	activeScreen = &MainScreen
}

func GetActiveScreen() *ui.Screen {
	return activeScreen
}

func SetActiveScreen(screenID ScreenID) {
	screen, ok := screenMap[screenID]
	if !ok {
		panic("invalid screen ID")
	}

	activeScreen = screen

	// FIXME: i dont think this can run if drawing has not begun so thats something to look out for
	rl.ClearBackground(screen.BackgroundColor)
}
