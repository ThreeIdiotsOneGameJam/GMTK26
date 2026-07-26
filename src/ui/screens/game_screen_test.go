package screens

import (
	"math"
	"testing"
	"time"
)

func TestRoundCountdownState(t *testing.T) {
	tests := []struct {
		name         string
		remaining    time.Duration
		wantDigit    int
		wantProgress float64
		wantVisible  bool
	}{
		{name: "before countdown", remaining: 3*time.Second + time.Nanosecond},
		{name: "three starts", remaining: 3 * time.Second, wantDigit: 3, wantVisible: true},
		{name: "three halfway", remaining: 2500 * time.Millisecond, wantDigit: 3, wantProgress: 0.5, wantVisible: true},
		{name: "two starts", remaining: 2 * time.Second, wantDigit: 2, wantVisible: true},
		{name: "two halfway", remaining: 1500 * time.Millisecond, wantDigit: 2, wantProgress: 0.5, wantVisible: true},
		{name: "one starts", remaining: time.Second, wantDigit: 1, wantVisible: true},
		{name: "one nearly ends", remaining: time.Nanosecond, wantDigit: 1, wantProgress: 1, wantVisible: true},
		{name: "deadline", remaining: 0},
		{name: "after deadline", remaining: -time.Nanosecond},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			digit, progress, visible := roundCountdownState(test.remaining)
			if digit != test.wantDigit {
				t.Fatalf("digit = %d, want %d", digit, test.wantDigit)
			}
			if visible != test.wantVisible {
				t.Fatalf("visible = %t, want %t", visible, test.wantVisible)
			}
			if math.Abs(progress-test.wantProgress) > 1e-8 {
				t.Fatalf("progress = %f, want %f", progress, test.wantProgress)
			}
		})
	}
}

func TestCountdownSecondsRemainingMatchesDisplayedDigit(t *testing.T) {
	tests := []struct {
		remaining time.Duration
		want      int
	}{
		{remaining: 3*time.Second + time.Nanosecond, want: 4},
		{remaining: 3 * time.Second, want: 3},
		{remaining: 2500 * time.Millisecond, want: 3},
		{remaining: 2 * time.Second, want: 2},
		{remaining: 1500 * time.Millisecond, want: 2},
		{remaining: time.Second, want: 1},
		{remaining: time.Nanosecond, want: 1},
		{remaining: 0, want: 0},
		{remaining: -time.Nanosecond, want: 0},
	}

	for _, test := range tests {
		if got := countdownSecondsRemaining(test.remaining); got != test.want {
			t.Errorf(
				"countdownSecondsRemaining(%s) = %d, want %d",
				test.remaining,
				got,
				test.want,
			)
		}
	}
}

func TestRoundAnnouncementState(t *testing.T) {
	tests := []struct {
		name        string
		round       int32
		remaining   time.Duration
		wantText    string
		wantVisible bool
	}{
		{
			name:        "new round",
			round:       2,
			remaining:   time.Second,
			wantText:    "Round #2",
			wantVisible: true,
		},
		{name: "announcement elapsed", round: 2, remaining: 0},
		{name: "no announced round", remaining: time.Second},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			text, visible := roundAnnouncementState(test.round, test.remaining)
			if text != test.wantText {
				t.Fatalf("text = %q, want %q", text, test.wantText)
			}
			if visible != test.wantVisible {
				t.Fatalf("visible = %t, want %t", visible, test.wantVisible)
			}
		})
	}
}

func TestEaseInExpoStartsSlowAndAccelerates(t *testing.T) {
	firstHalfGrowth := easeInExpo(0.5) - easeInExpo(0)
	secondHalfGrowth := easeInExpo(1) - easeInExpo(0.5)

	if firstHalfGrowth >= secondHalfGrowth {
		t.Fatalf(
			"expected accelerating growth, first half = %f, second half = %f",
			firstHalfGrowth,
			secondHalfGrowth,
		)
	}
}
