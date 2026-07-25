package audio

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/threeidiotsonegamejam/gmtk26/src/storage"
)

// Volumes set to 0 because I dont want to jumpscare you guys
var SFXVolume float32 = 0.0
var MusicVolume float32 = 0.0
var AmbienceVolume float32 = 0.0
var AmbienceVolumeMulti float32 = 1.0

var playingMusic = false
var playingAmbience = false

var song rl.Music
var ambience rl.Music

func Init() {
	rl.InitAudioDevice()

	storage.Load("sfx_volume", &SFXVolume)
	storage.Load("music_volume", &MusicVolume)
	storage.Load("ambience_volume", &AmbienceVolume)

	song = rl.LoadMusicStream("assets/audio/main_theme.ogg")
	ambience = rl.LoadMusicStream("assets/audio/ambience.ogg")
}

func Update() {
	if playingMusic {
		rl.SetMusicVolume(song, MusicVolume)
		rl.UpdateMusicStream(song)
	}
	if playingAmbience {
		rl.SetMusicVolume(ambience, AmbienceVolume*AmbienceVolumeMulti)
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
