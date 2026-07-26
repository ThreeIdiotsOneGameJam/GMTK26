package audio

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/global"
	"github.com/threeidiotsonegamejam/gmtk26/src/settings"
)

var AmbienceVolumeMulti float32 = 1.0

var playingMusic = false
var playingAmbience = false

var song rl.Music
var ambience rl.Music

func Init() {
	rl.InitAudioDevice()

	song = rl.LoadMusicStream(global.AssetDir + "/audio/main_theme.ogg")
	ambience = rl.LoadMusicStream(global.AssetDir + "/audio/ambience.ogg")
}

func Update() {
	if playingMusic {
		rl.SetMusicVolume(song, settings.Current.MusicVolume)
		rl.UpdateMusicStream(song)
	}
	if playingAmbience {
		rl.SetMusicVolume(ambience, settings.Current.AmbienceVolume*AmbienceVolumeMulti)
		rl.UpdateMusicStream(ambience)
	}
}

func Terminate() {
	rl.UnloadMusicStream(song)
	rl.UnloadMusicStream(ambience)
	rl.CloseAudioDevice()
}

func StartMusic() {
	rl.PlayMusicStream(song)
	playingMusic = true
}

func StopMusic() {
	rl.StopMusicStream(song)
	playingMusic = false
}

func StartAmbience() {
	rl.PlayMusicStream(ambience)
	playingAmbience = true
}

func StopAmbience() {
	rl.StopMusicStream(ambience)
	playingAmbience = false
}
