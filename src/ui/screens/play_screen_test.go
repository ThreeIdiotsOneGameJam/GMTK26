package screens

import (
	"testing"
	"time"
)

func TestPlayErrorExpiresAndRestoresHelper(t *testing.T) {
	SetPlayError("game not found")
	t.Cleanup(func() {
		SetPlayError("")
	})

	if playError != "Game not found" {
		t.Fatalf("playError = %q, want %q", playError, "Game not found")
	}

	advancePlayError(playErrorDisplayDuration - time.Nanosecond)
	if playError == "" {
		t.Fatal("play error expired before its display duration")
	}

	advancePlayError(time.Nanosecond)
	if playError != "" {
		t.Fatalf("playError = %q after display duration, want empty", playError)
	}
}

func TestSettingPlayErrorRestartsDisplayDuration(t *testing.T) {
	SetPlayError("first error")
	t.Cleanup(func() {
		SetPlayError("")
	})

	advancePlayError(playErrorDisplayDuration / 2)
	SetPlayError("second error")
	advancePlayError(playErrorDisplayDuration / 2)

	if playError != "Second error" {
		t.Fatalf("playError = %q, want the replacement error to remain visible", playError)
	}
	if playErrorTimeRemaining != playErrorDisplayDuration/2 {
		t.Fatalf(
			"playErrorTimeRemaining = %s, want %s",
			playErrorTimeRemaining,
			playErrorDisplayDuration/2,
		)
	}
}
