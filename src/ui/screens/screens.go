package screens

import (
	"github.com/threeidiotsonegamejam/gmtk26/src/ui"
)

type ScreenID int

const (
	MainScreenID ScreenID = iota
	PlayScreenID
	GameScreenID
	SettingsScreenID
)

var screenMap map[ScreenID]*ui.ScreenElement
var activeScreen *ui.ScreenElement

func init() {
	screenMap = map[ScreenID]*ui.ScreenElement{
		MainScreenID:     MainScreen,
		PlayScreenID:     PlayScreen,
		GameScreenID:     GameScreen,
		SettingsScreenID: SettingsScreen,
	}
	activeScreen = MainScreen
}

func GetActiveScreen() *ui.ScreenElement {
	return activeScreen
}

// FIXME: ideally this should schedule the screen change for the next frame instead of doing it immediately but its not a big deal rn
func SetActiveScreen(screenID ScreenID) {
	screen, ok := screenMap[screenID]
	if !ok {
		panic("invalid screen ID")
	}

	activeScreen = screen
}
